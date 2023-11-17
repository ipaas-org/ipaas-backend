package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// todo: is not available when there is a web applpication with the same name
// todo: skip this step if it is a database
func (c *Controller) IsNameAvailableSystemWide(ctx context.Context, name string) bool {
	_, err := c.ApplicationRepo.FindByName(ctx, name)
	available := err == repo.ErrNotFound
	c.l.Debugf("is name[%s] system available: %t", name, available)
	return available
}

// todo: use this function to check if the name is available for a database
func (c *Controller) IsNameAvailableUserWide(ctx context.Context, name, username string) bool {
	_, err := c.ApplicationRepo.FindByNameAndOwner(ctx, name, username)
	available := err == repo.ErrNotFound
	c.l.Debugf("is name[%s] available for %s: %t", name, username, available)
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

func (c *Controller) CreateApplicationFromApplicationIDandImageID(ctx context.Context, applicationID, imageID, BuildCommit string) error {
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
	//update status of application to starting and set the built commit to builtCommit
	app.State = model.ApplicationStateStarting
	app.BuiltCommit = BuildCommit
	if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
		c.l.Errorf("error updating application: %v", err)
		return err
	}

	//get user from the application owner and get his docker network id
	user, err := c.UserRepo.FindByCode(ctx, app.Owner)
	if err != nil {
		c.l.Errorf("error user not found for application owner: %v", err)
		return err
	}

	//start a container from the application
	labels := c.GenerateLabels(app.Name, app.Owner, app.Kind)
	host := fmt.Sprintf("%s.localhost", app.Name)
	traefikLabels := c.GenerateTraefikDnsLables(app.Name, host, app.ListeningPort)
	labels = append(labels, traefikLabels...)
	container, err := c.serviceManager.CreateNewContainer(ctx, app.Name, imageID, app.Envs, labels)
	if err != nil {
		c.l.Errorf("error creating new container: %v", err)
		app.State = model.ApplicationStateFailed
		if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
			c.l.Errorf("error updating application: %v", err)
		}
		return err
	}
	//connect container to user's network id and set as dns the application name
	c.serviceManager.ConnectContainerToNetwork(ctx, container.ContainerID, user.NetworkID, app.Name)
	//todo: add traefik network id (it should be in the config file for now, in the future it will be a dns service)
	//for now traefik is connected to host network so it's fine

	//start the container
	if err := c.serviceManager.StartContainerByID(ctx, container.ContainerID); err != nil {
		c.l.Errorf("error starting container: %v", err)
		app.State = model.ApplicationStateFailed
		if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
			c.l.Errorf("error updating application: %v", err)
		}
		return err
	}

	//update the application with the container informations and status running
	app.State = model.ApplicationStateRunning
	app.Container = container
	app.DnsName = host
	if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
		c.l.Errorf("error updating application: %v", err)
		return err
	}
	return nil
}
