package controller

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
)

func (c *Controller) BuildImage(ctx context.Context, app *model.Application, providerToken string) error {
	if app.State == model.ApplicationStateBuilding ||
		app.State == model.ApplicationStateStarting {
		return ErrInvalidOperationInCurrentState
	}

	request := model.BuildRequest{
		ApplicationID: app.ID.Hex(),
		Token:         providerToken,
		UserID:        app.Owner,
		Type:          "repo",
		Connector:     "github",
		Repo:          app.GithubRepo,
		Branch:        app.GithubBranch,
	}

	c.l.Debugf("sending to rmq: %+v ", request)
	if err := c.imageBuilder.BuildImage(request); err != nil {
		c.l.Errorf("error sending image to image builder: %v", err)
		app.State = model.ApplicationStateFailed
		if err := c.updateApplication(ctx, app); err != nil {
			return err
		}
		return err
	}

	return nil
}
