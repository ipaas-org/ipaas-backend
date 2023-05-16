package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type serviceType string

const (
	WebType      serviceType = "web"
	DatabaseType serviceType = "database"

	StatusCreating = "creating"
	StatusRunning  = "running"
)

// todo: is not available when there is a web applpication with the same name
// todo: skip this step if it is a database
func (c *Controller) IsNameAvailableSystemWide(ctx context.Context, name string) bool {
	_, err := c.applicationRepo.FindByName(ctx, name)
	c.l.Debugf("checking if name is available: %s", err.Error())
	return err == nil || err == repo.ErrNotFound
}

// todo: use this function to check if the name is available for a database
func (c *Controller) IsNameAvailableUserWide(ctx context.Context, name, username string) bool {
	_, err := c.applicationRepo.FindByNameAndOwnerUsername(ctx, name, username)
	return err == nil
}

func (c *Controller) CreateNewContainer(ctx context.Context, serviceType serviceType, ownerID, name, image string, env []model.KeyValue) (string, string, error) {
	application := &model.Application{
		Name:          name,
		OwnerUsername: ownerID,
		Status:        StatusCreating,
		Type:          string(serviceType),
		CreatedAt:     time.Now(),
		Envs:          env,
	}

	_, err := c.applicationRepo.Insert(ctx, application)
	if err != nil {
		return "", "", err
	}

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

func (c *Controller) CreateContainerFromIDAndImage(ctx context.Context, id, image string) error {

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

	containerID, _, err := c.CreateNewContainer(ctx, WebType, app.OwnerUsername, app.Name, app.ImageID, app.Envs)
	if err != nil {
		c.l.Errorf("error creating new container: %v", err)
		return fmt.Errorf("c.CreateNewContainer: %w", err)
	}

	app.Status = StatusRunning
	app.ImageID = image
	app.ContainerID = containerID

	if _, err := c.applicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
		c.l.Errorf("error updating application status: %v", err)
		return fmt.Errorf("c.applicationRepo.UpdateByID: %w", err)
	}
	return c.StartContainer(ctx, containerID)
}
