package controller

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (c *Controller) IsNameAvailableSystemWide(ctx context.Context, name string) bool {
	_, err := c.ApplicationRepo.FindByName(ctx, name)
	available := err == repo.ErrNotFound
	c.l.Debugf("is name[%s] system available: %t", name, available)
	return available
}

func (c *Controller) DoesApplicationExists(ctx context.Context, applicationID primitive.ObjectID) (bool, error) {
	_, err := c.ApplicationRepo.FindByID(ctx, applicationID)
	if err != nil {
		if err == repo.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *Controller) GetApplicationByID(ctx context.Context, id primitive.ObjectID) (*model.Application, error) {
	return c.ApplicationRepo.FindByID(ctx, id)
}

func (c *Controller) GetAllUserApplications(ctx context.Context, userCode string) ([]*model.Application, error) {
	return c.ApplicationRepo.FindByOwner(ctx, userCode)
}

// this function returns all the application of the user given the kind of the application
func (c *Controller) GetApplicationByKind(ctx context.Context, userCode string, kind model.ApplicationKind) ([]*model.Application, error) {
	return c.ApplicationRepo.FindByOwnerAndKind(ctx, userCode, kind)
}

func (c *Controller) InsertApplication(ctx context.Context, app *model.Application) error {
	app.ID = primitive.NewObjectID()
	_, err := c.ApplicationRepo.InsertOne(ctx, app)
	if err != nil {
		return fmt.Errorf("error inserting application: %v", err)
	}
	return nil
}

func (c *Controller) updateApplication(ctx context.Context, app *model.Application) error {
	if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
		c.l.Errorf("error updating application: %v", err)
		return err
	}
	return nil
}

// this function will insert a new application and send the build request to image builder
func (c *Controller) CreateNewWebApplication(ctx context.Context, userCode, providerAccessToken, name, gitRepo, gitBranch, listeningPort string, envs []model.KeyValue) (*model.Application, error) {
	app := new(model.Application)
	app.Name = name
	app.Kind = model.ApplicationKindWeb
	app.State = model.ApplicationStatePending
	app.CreatedAt = time.Now()
	app.Owner = userCode
	//todo: allow user to set visibility
	//! for now all applications are public
	app.Visiblity = model.ApplicationVisiblityPublic
	app.IsUpdatable = false
	app.ListeningPort = listeningPort
	app.GithubBranch = gitBranch
	app.GithubRepo = gitRepo
	app.Envs = envs

	if err := c.InsertApplication(ctx, app); err != nil {
		c.l.Errorf("error inserting application: %v", err)
		return nil, err
	}

	if err := c.BuildImage(ctx, app, providerAccessToken); err != nil {
		c.l.Errorf("error building image: %v", err)
		return nil, err
	}
	return app, nil
}

func (c *Controller) CreateApplicationFromApplicationIDandImageID(ctx context.Context, applicationID, registryImage, BuildCommit string) error {
	//convert the app id to primitive object id
	appID, err := primitive.ObjectIDFromHex(applicationID)
	if err != nil {
		c.l.Errorf("error converting application id to primitive, this error should not happen: %v", err)
		return err
	}

	//get application from application id
	app, err := c.ApplicationRepo.FindByID(ctx, appID)
	if err != nil {
		//!if this error happens it means that the application was not created succesfully or was deleted during the build process
		//todo: it should not happen, the user can not delete the application while building for now
		//todo: when implemented correctly the user will be able to stop the building process and delete the application
		c.l.Errorf("error getting application from application id: %v", err)
		return err
	}

	//get user from the application owner and get his docker network id
	user, err := c.UserRepo.FindByCode(ctx, app.Owner)
	if err != nil {
		c.l.Errorf("error user not found for application owner: %v", err)
		return err
	}

	c.l.Infof("create application deployment for %s[%s]", app.Name, app.ID.Hex())
	//update status of application to starting and set the built commit to builtCommit
	app.State = model.ApplicationStateStarting
	if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
		c.l.Errorf("error updating application: %v", err)
		return err
	}

	app.BuiltCommit = BuildCommit
	if app.Service != nil {
		c.l.Infof("updating deployment with new image")

		deployment, err := c.ServiceManager.UpdateDeployment(ctx, user.Namespace, app.Service.Deployment.Name, registryImage, app.Service.Deployment.Replicas, app.Service.Deployment.Port, app.Service.Deployment.Labels, "")
		if err != nil {
			c.l.Errorf("error updating deployment: %v", err)
			return err
		}
		deployment.ConfigMap = app.Service.Deployment.ConfigMap
		deployment.Volume = app.Service.Deployment.Volume
		app.Service.Deployment = deployment
		if err := c.updateApplication(ctx, app); err != nil {
			return err
		}
		return nil
	}

	configMap, err := c.createConfigMap(ctx, app, user, app.Envs)
	if err != nil {
		return err
	}

	//todo: implement volumes
	deployment, err := c.createDeployment(ctx, app, user, registryImage, configMap.Name, nil)
	if err != nil {
		return err
	}
	deployment.ConfigMap = configMap

	//watch for deployment ready state, it's non blocking so we just check at the end
	done, errChan := c.ServiceManager.WaitDeploymentReadyState(ctx, user.Namespace, deployment.Name)
	var errWhileWaiting error
loop:
	for {
		select {
		case err := <-errChan:
			c.l.Errorf("error while waiting for deployment: %v", err)
			errWhileWaiting = err
			break loop
		case _, ok := <-done:
			if !ok {
				//todo: choose what to do in this case
				c.l.Errorf("done chan is closed, either context was cancelled or something worst O-O")
				if ctx.Err() != nil {
					c.l.Errorf("context cancelled while a deployment was being created, no further changes will be done at the moment")
				} else {
					errWhileWaiting = fmt.Errorf("internal error")
				}
			} else {
				c.l.Debugf("deployment %s is available", deployment.Name)
			}
			break loop
		}
	}

	service, err := c.createService(ctx, app, user)
	if err != nil {
		return err
	}
	service.Deployment = deployment

	host := fmt.Sprintf("%s.%s", app.Name, c.app.BaseDefaultDomain)
	ingressRoute, err := c.createIngressRoute(ctx, app, user, host, service.Name, service.Port)
	if err != nil {
		return err
	}
	service.IngressRoute = ingressRoute

	if errWhileWaiting != nil {
		//todo: handle waiting error, it's probably because it reached a timeout
		//in this case it probably means that we reached a cpu/mem cap and we should
		//expand the infra, it should really not happen
		c.l.Errorf("internal error reached, this should not happen, check infrastructure resources left")
		app.State = model.ApplicationStateFailed
	} else {
		app.State = model.ApplicationStateRunning
		app.Service = service
		app.DnsName = host
	}

	//update the application with the container informations and status running
	if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
		c.l.Errorf("error updating application: %v", err)
		return err
	}

	return nil
}

func (c *Controller) DeleteApplication(ctx context.Context, app *model.Application, user *model.User) error {
	//! this version is not able to delete a pending build cause it's unable to delete a message
	// in the rmq queue, in the next version it will not be a problem cause we will also be able to stop the building process
	if app.State == model.ApplicationStateBuilding ||
		app.State == model.ApplicationStateStarting {
		return ErrInvalidOperationInCurrentState
	}

	app.State = model.ApplicationStateDeleting
	if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
		c.l.Errorf("error updating application: %v", err)
		return err
	}

	if err := c.deleteDeployment(ctx, app, user); err != nil {
		return err
	}

	if err := c.deleteService(ctx, app, user); err != nil {
		return err
	}

	if err := c.deleteIngressRoute(ctx, app, user); err != nil {
		return err
	}

	if err := c.deletePersistantVolumeClmain(ctx, app, user); err != nil {
		return err
	}

	if err := c.deleteConfigMap(ctx, app, user); err != nil {
		return err
	}

	return nil
}

func (c *Controller) FailedBuild(ctx context.Context, info *model.BuildResponse) error {
	appID, err := primitive.ObjectIDFromHex(info.ApplicationID)
	if err != nil {
		return err
	}

	app, err := c.ApplicationRepo.FindByID(ctx, appID)
	if err != nil {
		return err
	}

	app.BuiltCommit = info.BuiltCommit
	app.State = model.ApplicationStateFailed
	//TODO: write info.Message in the app logs
	if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
		return err
	}
	return nil
}

func (c *Controller) RedeployApplication(ctx context.Context, user *model.User, application *model.Application) error {
	c.l.Infof("force restart of deployment %s (appID=%s) of user %s", application.Service.Deployment.Name, application.ID.Hex(), user.Code)
	application.Service.Deployment.CurrentPodName = ""
	if err := c.updateApplication(ctx, application); err != nil {
		return err
	}
	if err := c.ServiceManager.RestartDeployment(ctx, user.Namespace, application.Service.Deployment.Name); err != nil {
		c.l.Errorf("error restarting deployment %s: %v", application.Service.Deployment.Name, err)
		return err
	}
	return nil
}

func (c *Controller) AddCurrentPodToApplication(ctx context.Context, applicationID primitive.ObjectID, podName string) {
	go func() {
		for {
			app, err := c.ApplicationRepo.FindByID(ctx, applicationID)
			if err != nil {
				c.l.Errorf("error finding application by id: %v", err)
				return
			}
			if app.Service == nil || app.Service.Deployment == nil {
				c.l.Debugf("application still not fully created, waiting to add pod")
				time.Sleep(50 * time.Millisecond)
				continue
			}
			app.Service.Deployment.CurrentPodName = podName
			if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
				c.l.Errorf("error updating application: %v", err)
			}
			c.l.Infof("pod %s is now the current pod for application %s", podName, applicationID.Hex())
			return
		}
	}()
}

func (c *Controller) UpdateApplication(ctx context.Context, app *model.Application, user *model.User, name string, patchPort string, envs []model.KeyValue) error {
	c.l.Debugf("updating application %s", app.ID.Hex())
	fields := make(logrus.Fields)
	fields["applicationID"] = app.ID.Hex()
	fields["userID"] = user.Code
	fields["action"] = "UpdateApplication"
	hasSomethingChanged := false
	if name != "" && app.Name != name {
		hasSomethingChanged = true
		fields["name"] = name
		fields["oldName"] = app.Name
		c.l.WithFields(fields).Info("user trying to change name, currently not implemented")
		return ErrInvalidOperationInCurrentState
		// switch app.Kind {
		// case model.ApplicationKindWeb, model.ApplicationKindManagment:
		// 	if c.IsNameAvailableSystemWide(ctx, name) {
		// 		return ErrApplicationNameNotAvailable
		// 	}
		// 	c.ServiceManager.UpdateIngressRoute()
		// case model.ApplicationKindStorage:
		// 	return ErrInvalidOperationInCurrentState
		// }
		// app.Name = name
	}

	if patchPort != "" && app.ListeningPort != patchPort {
		hasSomethingChanged = true
		fields["port"] = patchPort
		fields["oldPort"] = app.ListeningPort
		port, err := strconv.Atoi(patchPort)
		if err != nil {
			c.l.WithFields(fields).Info("user trying to change port, invalid port")
			return ErrInvalidPort
		}
		if port < 0 || port > 65535 {
			c.l.WithFields(fields).Info("user trying to change port, invalid port")
			return ErrInvalidPort
		}
		app.ListeningPort = patchPort
		updatedService, err := c.ServiceManager.UpdateService(ctx, user.Namespace, app.Service.Name, int32(port))
		if err != nil {
			c.l.WithFields(fields).Errorf("error updating service: %v", err)
			return err
		}
		updatedService.Deployment = app.Service.Deployment
		updatedService.IngressRoute = app.Service.IngressRoute
		app.Service = updatedService
		c.l.WithFields(fields).Debugf("service %s updated succesfully with new port", app.Service.Name)
	}

	configMapName := ""
	if envs != nil {
		// fields["envs"]=
		for _, env := range envs {
			c.l.Debugf("env: %+v", env)
			if env.Key == "" || env.Value == "" {
				return ErrInvalidEnv
			}
		}
		same := true
		if len(envs) != len(app.Envs) {
			same = false
		} else {
			oldEnvMap := convertModelKeyValueToMap(app.Envs)
			newEnvMap := convertModelKeyValueToMap(envs)
			for key, value := range oldEnvMap {
				if newEnvMap[key] != value {
					same = false
					break
				}
			}
		}
		if same {
			c.l.WithFields(fields).Info("user trying to update envs, but they are the same")
		} else {
			hasSomethingChanged = true
			app.Envs = envs
			if app.Service.Deployment.ConfigMap != nil {
				updatedConfigMap, err := c.ServiceManager.UpdateConfigMap(ctx, user.Namespace, app.Service.Deployment.ConfigMap.Name, envs)
				if err != nil {
					c.l.WithFields(fields).Errorf("error updating config map %s: %v", app.Service.Deployment.ConfigMap.Name, err)
					return err
				}
				app.Service.Deployment.ConfigMap = updatedConfigMap
				c.l.WithFields(fields).Debugf("configMap %s updated succesfully", app.Service.Deployment.ConfigMap.Name)
			} else {
				configMap, err := c.createConfigMap(ctx, app, user, envs)
				if err != nil {
					c.l.WithFields(fields).Errorf("error creating config map: %v", err)
					return err
				}
				app.Service.Deployment.ConfigMap = configMap
				c.l.WithFields(fields).Debugf("configMap %s created succesfully", configMap.Name)
			}
			configMapName = app.Service.Deployment.ConfigMap.Name
		}
	}

	if !hasSomethingChanged {
		return ErrNoChanges
	}

	updatedDeployment, err := c.ServiceManager.UpdateDeployment(
		ctx,
		user.Namespace,
		app.Service.Deployment.Name,
		app.Service.Deployment.ImageRegistry,
		app.Service.Deployment.Replicas,
		app.Service.Deployment.Port,
		app.Service.Deployment.Labels,
		configMapName)
	if err != nil {
		c.l.WithFields(fields).Errorf("error updating deployment %s: %v", app.Service.Deployment.Name, err)
		return err
	}
	c.l.WithFields(fields).Debugf("deployment %s updated succesfully", app.Service.Deployment.Name)
	if configMapName != "" {
		updatedDeployment.ConfigMap = app.Service.Deployment.ConfigMap
	}
	updatedDeployment.Volume = app.Service.Deployment.Volume
	app.Service.Deployment = updatedDeployment

	//redeploy application
	c.l.WithFields(fields).Debugf("redeploy application to apply changes")
	if err := c.RedeployApplication(ctx, user, app); err != nil {
		c.l.WithFields(fields).Errorf("error redeploy application during update")
		return err
	}

	if err := c.updateApplication(ctx, app); err != nil {
		return err
	}
	c.l.WithFields(fields).Infof("application updated succesfully")
	return nil
}

func (c *Controller) RolloutApplication(ctx context.Context, user *model.User, app *model.Application) error {
	c.l.Infof("rollout of deployment %s (appID=%s) of user %s", app.Service.Deployment.Name, app.ID.Hex(), user.Code)
	if app.Kind != model.ApplicationKindWeb {
		return ErrInvalidOperationWithCurrentKind
	}

	if app.State != model.ApplicationStateRunning &&
		app.State != model.ApplicationStateFailed &&
		app.State != model.ApplicationStateCrashed &&
		app.State != model.ApplicationStateRollingOut {
		return ErrInvalidOperationInCurrentState
	}

	newLastHash, err := c.GetLastCommitHash(ctx, user, app.GithubRepo, app.GithubBranch)
	if err != nil {
		c.l.Errorf("error getting last commit hash: %v", err)
		return err
	}
	if newLastHash == app.BuiltCommit {
		c.l.Infof("no new commit to rollout")
		return ErrLastVersionAlreadyDeployed
	}

	app.State = model.ApplicationStateRollingOut
	if err := c.updateApplication(ctx, app); err != nil {
		return err
	}

	if err := c.BuildImage(ctx, app, user.Info.GithubAccessToken); err != nil {
		c.l.Errorf("error building image: %v", err)
		return err
	}
	return nil
}
