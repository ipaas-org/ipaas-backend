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
func (c *Controller) GetApplicationByKind(ctx context.Context, userCode string, kind model.ServiceKind) ([]*model.Application, error) {
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

	//get user from the application owner and get his docker network id
	user, err := c.UserRepo.FindByCode(ctx, app.Owner)
	if err != nil {
		c.l.Errorf("error user not found for application owner: %v", err)
		return err
	}

	c.l.Infof("create the container for %s[%s]", app.Name, app.ID)
	//update status of application to starting and set the built commit to builtCommit
	app.State = model.ApplicationStateStarting
	app.BuiltCommit = BuildCommit
	if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
		c.l.Errorf("error updating application: %v", err)
		return err
	}
	//start a container from the application
	labels := c.GenerateLabels(app.Name, app.Owner, app.Kind)
	host := fmt.Sprintf("%s.%s", app.Name, c.app.BaseDefaultDomain)
	traefikLabels := c.GenerateTraefikDnsLables(app.Name, host, app.ListeningPort)
	labels = append(labels, traefikLabels...)

	container, err := c.createConnectAndStartContainer(ctx, app.Name, imageID, user.NetworkID, app.Envs, labels)
	if err != nil {
		c.l.Errorf("error creating container: %v", err)
		app.State = model.ApplicationStateFailed
		if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
			c.l.Errorf("error updating application: %v", err)
		}
		return err
	}

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
	app.Container = container
	app.DnsName = host
	if _, err := c.ApplicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
		c.l.Errorf("error updating application: %v", err)
		return err
	}

	return nil
}

// TODO: this should be a service, like the label generator for traefik
// func (c *Controller) checkIfTraefikUpdated(ctx context.Context, name string, retry int) (bool, error) {
// 	//<ip>:<port>/api/http/routers/<name>@docker
// 	//returns 404 if traefik didnt update yet
// 	//200 after traefik updates
// 	url := fmt.Sprintf("%s/api/http/routers/%s@docker", c.traefik.ApiBaseUrl, name)
// 	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
// 	if err != nil {
// 		return false, err
// 	}

// 	auth := c.traefik.Username + ":" + c.traefik.Password
// 	base64Auth := base64.StdEncoding.EncodeToString([]byte(auth))
// 	req.Header.Add("Authorization", "Basic "+base64Auth)

// 	client := new(http.Client)
// 	counter := 0
// 	for {
// 		c.l.Debugf("doing request to traefik for %s@docker service", name)
// 		resp, err := client.Do(req)
// 		if err != nil {
// 			c.l.Errorf("error doing request to traefik: %v", err)
// 			//internal server error
// 			return false, err
// 		}
// 		body, _ := io.ReadAll(resp.Body)
// 		c.l.Debugf("status %d, body: %s", resp.StatusCode, string(body))
// 		if resp.StatusCode == 200 {
// 			c.l.Debugf("traefik updated successfully for %s@docker service", name)
// 			break
// 		}
// 		if resp.StatusCode != 404 {
// 			//internal server error
// 			return false, fmt.Errorf("error doing request to traefik: %d %s", resp.StatusCode, body)
// 		}
// 		if counter >= retry {
// 			//internal server error
// 			c.l.Debug("traefik did not update in time")
// 			return false, nil
// 		}
// 		c.l.Debugf("traefik did not update yet (current try %d), retrying in 5 second", counter)
// 		counter++
// 		time.Sleep(5 * time.Second)
// 		continue
// 	}
// 	return true, nil
// }

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
	if application.Container != nil {
		if err := c.serviceManager.StopContainerByID(ctx, application.Container.ID); err != nil {
			c.l.Errorf("error stopping container[%s] of user %s for application %s: %v", application.Container.ID, application.Owner, application.ID.Hex(), err)
			return err
		}

		if err := c.serviceManager.RemoveContainerByID(ctx, application.Container.ID, false); err != nil {
			c.l.Errorf("error removing container[%s] of user %s for application %s: %v", application.Container.ID, application.Owner, application.ID.Hex(), err)
			return err
		}

		if application.Kind == model.ApplicationKindWeb {
			//delete the image
			if err := c.serviceManager.RemoveImageByID(ctx, application.Container.ImageID); err != nil {
				c.l.Errorf("error removing image[%s] of user %s for applicationg %s: %v", application.Container.ImageID, application.Owner, application.ID.Hex(), err)
				return err
			}
		}
	}

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
