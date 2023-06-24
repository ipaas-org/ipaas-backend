package controller

import (
	"context"
	"fmt"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	StatusCreated  = "created"
	StatusPending  = "pending"
	StatusStarting = "starting"
	StatusRunning  = "running"
	StatusFailed   = "failed"
	StatusBuilding = "building"
	StatusUpdating = "updating"
)

var (
	ErrNameNotAvailable = fmt.Errorf("name is not available")
)

// todo: is not available when there is a web applpication with the same name
// todo: skip this step if it is a database
func (c *Controller) IsNameAvailableSystemWide(ctx context.Context, name string) bool {
	_, err := c.applicationRepo.FindByName(ctx, name)
	available := err == repo.ErrNotFound
	c.l.Debugf("is name[%s] system available: %t", name, available)
	return available
}

// todo: use this function to check if the name is available for a database
func (c *Controller) IsNameAvailableUserWide(ctx context.Context, name, username string) bool {
	_, err := c.applicationRepo.FindByNameAndOwnerUsername(ctx, name, username)
	available := err == repo.ErrNotFound
	c.l.Debugf("is name[%s] available for %s: %t", name, username, available)
	return available
}

func (c *Controller) DoesApplicationExists(app *model.Application) (bool, error) {
	_, err := c.applicationRepo.FindByID(context.Background(), app.ID)
	if err != nil {
		if err == repo.ErrNotFound {
			return false, nil
		}
		return false, fmt.Errorf("c.applicationRepo.FindByID: %w", err)
	}
	return true, nil
}

func (c *Controller) GetApplicationByID(ctx context.Context, id primitive.ObjectID) (*model.Application, error) {
	return c.applicationRepo.FindByID(ctx, id)
}

func (c *Controller) InsertApplication(ctx context.Context, app *model.Application) error {
	app.Status = StatusCreated
	id, err := c.applicationRepo.Insert(ctx, app)
	if err != nil {
		return fmt.Errorf("error inserting application: %v", err)
	}
	//convert id to string
	app.ID = id.(primitive.ObjectID)
	return nil
}

func (c *Controller) UpdateApplicationState(ctx context.Context, app *model.Application) error {
	if _, err := c.applicationRepo.UpdateByID(ctx, app, app.ID); err != nil {
		c.l.Errorf("error updating application status: %v", err)
		return fmt.Errorf("c.applicationRepo.UpdateByID: %w", err)
	}
	return nil
}

func (c *Controller) CreateNewApplication(ctx context.Context, app *model.Application, providerAccessToken string) error {
	if err := c.InsertApplication(ctx, app); err != nil {
		return fmt.Errorf("c.InsertApplication: %w", err)
	}

	if err := c.BuildImage(ctx, app, providerAccessToken); err != nil {
		return fmt.Errorf("c.BuildImage: %w", err)
	}
	return nil
}
