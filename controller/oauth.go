package controller

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
)

func (c *Controller) CreateState() (string, error) {
	ran, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return "st_" + ran.String(), nil
}

func (c *Controller) createTempTokenKey() (string, error) {
	ran, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return "tk_" + ran.String(), nil
}

func (c *Controller) InsertState(ctx context.Context, state string, kind model.StateKind) error {
	id, err := c.StateRepo.InsertOne(ctx, &model.State{
		State:          state,
		Kind:           kind,
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

func (c *Controller) CheckState(ctx context.Context, state string) (bool, model.StateKind, error) {
	s, err := c.StateRepo.FindByState(ctx, state)
	if err != nil {
		if err == repo.ErrNotFound {
			return false, "", nil
		} else {
			return false, "", err
		}
	}
	if _, err := c.StateRepo.DeleteByState(ctx, state); err != nil {
		return false, "", err
	}
	c.l.Debug("found state with kind:", s.Kind)
	return true, s.Kind, nil
}

func (c *Controller) GenerateLoginUri(ctx context.Context, kind model.StateKind) string {
	state, err := c.CreateState()
	if err != nil {
		c.l.Errorf("Error creating state: %s", err.Error())
		return ""
	}
	if err := c.InsertState(ctx, state, kind); err != nil {
		c.l.Errorf("Error storing state: %s", err.Error())
		return ""
	}

	return c.gitProvider.GenerateLoginRedirectUri(ctx, state)
}

func (c *Controller) GetUserInfoFromOauthCode(ctx context.Context, code string) (*model.UserInfo, error) {
	accessToken, err := c.gitProvider.GetAccessTokenFromCode(ctx, code)
	if err != nil {
		c.l.Errorf("Error getting access token from code: %s", err.Error())
		return nil, err
	}

	user, err := c.gitProvider.GetUserInfo(ctx, accessToken)
	if err != nil {
		c.l.Errorf("Error getting user info: %s", err.Error())
		return nil, err
	}

	return user, nil
}

func (c *Controller) StoreTokensGenerateRedirectUri(ctx context.Context, jwt *model.AccessToken, refresh *model.RefreshToken) (string, error) {
	key, err := c.createTempTokenKey()
	if err != nil {
		return "", err
	}

	if err := c.TempTokenRepo.InsertTokens(ctx, key, jwt, refresh); err != nil {
		c.l.Errorf("error inserting tokens in temp storage")
		return "", err
	}
	uri := c.app.FrontendUrl + "/login?key=" + key
	return uri, nil
}

func (c *Controller) GetTokensFromKeyAndDeleteKey(ctx context.Context, key string) (*model.AccessToken, *model.RefreshToken, error) {
	jwt, refresh, err := c.TempTokenRepo.FindByKey(ctx, key)
	if err != nil {
		c.l.Errorf("error getting tokens from temp storage: %v", err)
		return nil, nil, err
	}
	if err := c.TempTokenRepo.DeleteKey(ctx, key); err != nil {
		c.l.Errorf("error deleting key from temp storage: %v", err)
		return nil, nil, err
	}
	return jwt, refresh, nil
}
