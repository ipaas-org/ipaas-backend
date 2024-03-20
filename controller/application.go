package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
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
	app.BuiltCommit = BuildCommit
	if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
		c.l.Errorf("error updating application: %v", err)
		return err
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

func (c *Controller) DeleteApplication(ctx context.Context, application *model.Application) error {
	//! this version is not able to delete a pending build cause it's unable to delete a message
	// in the rmq queue, in the next version it will not be a problem cause we will also be able to stop the building process
	if application.State == model.ApplicationStateBuilding ||
		application.State == model.ApplicationStateStarting {
		return ErrInvalidOperationInCurrentState
	}

	application.State = model.ApplicationStateDeleting
	if _, err := c.ApplicationRepo.UpdateByID(ctx, application, application.ID); err != nil {
		c.l.Errorf("error updating application: %v", err)
		return err
	}

	if _, err := c.ApplicationRepo.DeleteByID(ctx, application.ID); err != nil {
		c.l.Errorf("error deleting application %s: %v", application.ID.Hex(), err)
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
