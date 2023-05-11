package controller

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ipaas-org/ipaas-backend/model"
)

func (c *Controller) createAndStoreState(ctx context.Context) (string, error) {
	ran, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	stateValue := ran.String()

	id, err := c.stateRepo.Insert(ctx, &model.State{
		State:          stateValue,
		Issued:         time.Now(),
		ExpirationDate: time.Now().Add(time.Minute * 5),
	})
	if err != nil {
		c.l.Error("Error inserting state: %s", err.Error())
		return "", err
	}

	c.l.Debug("Inserted state(%s) with id: %s", stateValue, id)
	return stateValue, err
}

func (c *Controller) GenerateLoginUri(ctx context.Context) string {
	state, err := c.createAndStoreState(ctx)
	if err != nil {
		c.l.Error("Error creating state: %s", err.Error())
		return ""
	}

	return c.oauthService.GenerateLoginRedirectUri(state)
}

func (c *Controller) CheckState(ctx context.Context, state string) (bool, error) {
	return c.stateRepo.DeleteByState(ctx, state)
}

func (c *Controller) GetUserFromOauthCode(ctx context.Context, code string) (model.User, error) {
	accessToken, err := c.oauthService.GetAccessTokenFromCode(code)
	if err != nil {
		c.l.Error("Error getting access token from code: %s", err.Error())
		return model.User{}, err
	}

	user, err := c.oauthService.GetUserInfo(accessToken)
	if err != nil {
		c.l.Error("Error getting user info: %s", err.Error())
		return model.User{}, err
	}

	return user, nil
}
