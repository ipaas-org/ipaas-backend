package controller

import (
	"context"
	"fmt"
	"strconv"
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

func (c *Controller) UpdateApplication(ctx context.Context, app *model.Application) error {
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
	app.IsPublic = true //! for now all applications are public
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

	c.l.Infof("create the container for %s[%s]", app.Name, app.ID.Hex())
	//update status of application to starting and set the built commit to builtCommit
	app.State = model.ApplicationStateStarting
	app.BuiltCommit = BuildCommit
	if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
		c.l.Errorf("error updating application: %v", err)
		return err
	}
	//start a container from the application
	// labels := c.GenerateLabels(app.Name, app.Owner, app.Kind)
	host := fmt.Sprintf("%s.%s", app.Name, c.app.BaseDefaultDomain)
	// traefikLabels := c.GenerateTraefikDnsLables(app.Name, host, app.ListeningPort)
	// labels = append(labels, traefikLabels...)

	p, err := strconv.Atoi(app.ListeningPort)
	if err != nil {
		c.l.Errorf("error converting port to int: %v", err)
		return err
	}
	intPort := int32(p)
	// container, err := c.createConnectAndStartContainer(ctx, app.Name, imageID, user.Namespace, app.Envs, labels)
	deployment, err := c.serviceManager.CreateDeployment(ctx, user.Namespace, "deploy-"+app.Name, app.Name, user.Code, staticTempEnvironment, "public", registryImage, 1, intPort, app.Envs)
	if err != nil {
		c.l.Errorf("error creating deployment: %v", err)
		app.State = model.ApplicationStateFailed
		if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
			c.l.Errorf("error updating application: %v", err)
		}
		return err
	}

	service, err := c.serviceManager.CreateService(ctx, user.Namespace, "svc-"+app.Name, app.Name, user.Code, staticTempEnvironment, "public", intPort)
	if err != nil {
		c.l.Errorf("error creating service: %v", err)
		app.State = model.ApplicationStateFailed
		if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
			c.l.Errorf("error updating application: %v", err)
		}
		return err
	}
	service.Deployment = deployment

	ingressRoute, err := c.serviceManager.CreateIngressRoute(ctx, user.Namespace, "ingressroute-"+app.Name, app.Name, user.Code, staticTempEnvironment, "public", host, "svc-"+app.Name)
	if err != nil {
		c.l.Errorf("error creating ingress route: %v", err)
		app.State = model.ApplicationStateFailed
		if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
			c.l.Errorf("error updating application: %v", err)
		}
		return err
	}
	service.IngressRoute = ingressRoute

	//todo: this works but idk if it's timing based or there is a better way of doing it
	//todo: so for now we will just tell the user to wait a few seconds for our dns to update
	// updated, err := c.checkIfTraefikUpdated(ctx, app.Name, 15)
	// if err != nil {
	// 	return err
	// }
	// if !updated {
	// 	//! this should not happen, if it does check if traefik is healty
	// 	return fmt.Errorf("dns not updated successfully")
	// }

	//update the application with the container informations and status running
	app.State = model.ApplicationStateRunning
	app.Service = service
	app.DnsName = host
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

	//delete the container
	// if application.Implementation != nil {
	// 	if err := c.serviceManager.StopServiceByID(ctx, application.Implementation.ID); err != nil {
	// 		c.l.Errorf("error stopping container[%s] of user %s for application %s: %v", application.Implementation.ID, application.Owner, application.ID.Hex(), err)
	// 		return err
	// 	}

	// 	if err := c.serviceManager.RemoveServiceByID(ctx, application.Implementation.ID, false); err != nil {
	// 		c.l.Errorf("error removing container[%s] of user %s for application %s: %v", application.Implementation.ID, application.Owner, application.ID.Hex(), err)
	// 		return err
	// 	}

	// 	if application.Kind == model.ApplicationKindWeb {
	// 		//delete the image
	// 		if err := c.serviceManager.RemoveImageByID(ctx, application.Implementation.ImageID); err != nil {
	// 			c.l.Errorf("error removing image[%s] of user %s for applicationg %s: %v", application.Implementation.ImageID, application.Owner, application.ID.Hex(), err)
	// 			return err
	// 		}
	// 	}
	// }

	//delete the application from the db
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
