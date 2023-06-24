package controller

import (
	"context"
	"fmt"

	"github.com/ipaas-org/ipaas-backend/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type serviceType string

const (
	WebType      serviceType = "web"
	DatabaseType serviceType = "database"

)

func (c *Controller) CreateNewContainer(ctx context.Context, serviceType serviceType, ownerID, name, image string, env []model.KeyValue) (string, string, error) {
	labes := []model.KeyValue{
		{Key: "org.ipaas.service.name", Value: name},
		{Key: "org.ipaas.service.owner", Value: ownerID},
		{Key: "org.ipaas.service.type", Value: string(serviceType)},
		{Key: "org.ipaas.version", Value: c.app.Version},
		{Key: "org.ipaas.name", Value: c.app.Name},
		{Key: "org.ipaas.deployment", Value: c.app.Deployment},
	}

	return c.serviceManager.CreateNewContainer(ctx, name, image, env, labes)
}

func (c *Controller) RemoveNetwork(ctx context.Context, id string) error {
	return c.serviceManager.RemoveNetwork(ctx, id)
}

func (c *Controller) CreateNewNetwork(ctx context.Context, name string) (string, error) {
	return c.serviceManager.CreateNewNetwork(ctx, name)
}

func (c *Controller) ConnectContainerToNetwork(ctx context.Context, containerID, networkID, dnsAlias string) error {
	return c.serviceManager.ConnectContainerToNetwork(ctx, containerID, networkID, dnsAlias)
}

func (c *Controller) RemoveContainer(ctx context.Context, id string) error {
	return c.serviceManager.RemoveContainer(ctx, id)
}

func (c *Controller) StartContainer(ctx context.Context, id string) error {
	return c.serviceManager.StartContainer(ctx, id)
}

func (c *Controller) CreateContainerFromIDAndImage(ctx context.Context, id, lastCommitHash, image string) error {
	uuid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.l.Errorf("error parsing uuid: %v", err)
		return fmt.Errorf("primitive.ObjectIDFromHex: %w", err)
	}

	app, err := c.applicationRepo.FindByID(ctx, uuid)
	if err != nil {
		c.l.Errorf("error finding application: %v", err)
		return fmt.Errorf("c.applicationRepo.FindByID: %w", err)
	}

	c.l.Debugf("creating new container for application: %+v", app)
	containerID, _, err := c.CreateNewContainer(ctx, WebType, app.OwnerUsername, app.Name, image, app.Envs)
	if err != nil {
		c.l.Errorf("error creating new container: %v", err)
		return fmt.Errorf("c.CreateNewContainer: %w", err)
	}

	app.Status = StatusRunning
	app.ImageID = image
	app.ContainerID = containerID
	app.Type = string(WebType)
	app.LastCommitHash = lastCommitHash

	c.l.Debugf("container created correctly, updating application status: %+v", app)
	if _, err := c.applicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
		c.l.Errorf("error updating application status: %v", err)
		return fmt.Errorf("c.applicationRepo.UpdateByID: %w", err)
	}
	c.l.Debug("application status updated correctly")
	c.l.Debug("starting container")
	if err := c.StartContainer(ctx, containerID); err != nil {
		app.Status = StatusFailed
		c.applicationRepo.UpdateByID(ctx, app, app.ID)
		c.l.Errorf("error starting container: %v", err)
		return fmt.Errorf("c.StartContainer: %w", err)
	}

	return nil
}
