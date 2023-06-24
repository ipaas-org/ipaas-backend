package controller

import (
	"context"
	"errors"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
)

var (
	ErrNetworkIDNotSet = errors.New("network id not set")
)

func (c *Controller) DoesUserExist(ctx context.Context, email string) (bool, error) {
	_, err := c.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if err == repo.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *Controller) CreateUser(ctx context.Context, user *model.User) error {
	c.l.Debugf("Creating user %s (name=%q email=%q)", user.Username, user.FullName, user.Email)
	// networkID, err := c.CreateNewNetwork(ctx, user.Username)
	// if err != nil {
	// 	c.l.Errorf("Error creating network for user %s (name=%q email=%q): %s", user.Username, user.FullName, user.Email, err.Error())
	// 	return err
	// }
	if user.NetworkID == "" {
		return ErrNetworkIDNotSet
	}

	if user.Role == "" {
		user.Role = model.RoleUser
	}
	user.UserSettings.Theme = "light"

	_, err := c.userRepo.InsertOne(ctx, user)
	if err != nil {
		c.l.Errorf("Error creating user %s (name=%q email=%q): %s", user.Username, user.FullName, user.Email, err.Error())
		return err
	}
	c.l.Infof("user %s (name=%q email=%q) created successfully", user.Username, user.FullName, user.Email)
	return nil
}

func (c *Controller) GetUserFromEmail(ctx context.Context, email string) (*model.User, error) {
	return c.userRepo.FindByEmail(ctx, email)
}

// TODO
func (c *Controller) UpdateUser(ctx context.Context, user *model.User) error {
	return nil
}

// todo: delete user, stop and delete all containers, delete all volumes
// delete network, remove associated tokens
func (c *Controller) DeleteUser(ctx context.Context, email string) (bool, error) {
	// user, err := c.userRepo.FindByEmail(ctx, email)
	// if err != nil {
	// 	return false, err
	// }

	// deleted, err := c.userRepo.DeleteByID(ctx, user.ID)
	// if err != nil {
	// 	return false, err
	// }
	// c.RemoveNetwork(ctx, user.NetworkID)
	return false, nil
}
