package controller

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
)

func (c *Controller) DoesUserExist(ctx context.Context, user model.User) (bool, error) {
	_, err := c.userRepo.FindByEmail(ctx, user.Email)
	if err != nil {
		if err == repo.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *Controller) CreateUser(ctx context.Context, user model.User) error {
	c.l.Debugf("Creating user: %+v", user)
	_, err := c.userRepo.InsertOne(ctx, &user)
	return err
}

func (c *Controller) GetUserFromEmail(ctx context.Context, email string) (*model.User, error) {
	return c.userRepo.FindByEmail(ctx, email)
}
