package controller

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ipaas-org/ipaas-backend/model"
)

func (c *Controller) CreateState() (string, error) {
	ran, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return ran.String(), nil
}

func (c *Controller) InsertState(ctx context.Context, state string) error {
	id, err := c.StateRepo.InsertOne(ctx, &model.State{
		State:          state,
		Issued:         time.Now(),
		ExpirationDate: time.Now().Add(time.Minute * 5),
	})
	if err != nil {
		c.l.Errorf("Error inserting state: %s", err.Error())
		return err
	}

	c.l.Debugf("Inserted state(%s) with id: %s", state, id)
	return err
}

func (c *Controller) CheckState(ctx context.Context, state string) (bool, error) {
	return c.StateRepo.DeleteByState(ctx, state)
}

func (c *Controller) GenerateLoginUri(ctx context.Context) string {
	state, err := c.CreateState()
	if err != nil {
		c.l.Errorf("Error creating state: %s", err.Error())
		return ""
	}
	if err := c.InsertState(ctx, state); err != nil {
		c.l.Errorf("Error storing state: %s", err.Error())
		return ""
	}

	return c.gitProvider.GenerateLoginRedirectUri(state)
}

func (c *Controller) GetUserInfoFromOauthCode(ctx context.Context, code string) (*model.UserInfo, error) {
	accessToken, err := c.gitProvider.GetAccessTokenFromCode(code)
	if err != nil {
		c.l.Errorf("Error getting access token from code: %s", err.Error())
		return nil, err
	}

	user, err := c.gitProvider.GetUserInfo(accessToken)
	if err != nil {
		c.l.Errorf("Error getting user info: %s", err.Error())
		return nil, err
	}

	return user, nil
}
