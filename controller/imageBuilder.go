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

	request := model.Request{
		ApplicationID: app.ID.Hex(),
		PullInfo: &model.PullInfoRequest{
			Token:     providerToken,
			UserID:    app.Owner,
			Connector: "github",
			Repo:      app.GithubRepo,
			Branch:    app.GithubBranch,
		},
		BuildPlan: app.BuildConfig,
	}

	c.l.Debugf("sending to rmq: %+v ", request)
	if err := c.imageBuilder.BuildImage(ctx, request); err != nil {
		c.l.Errorf("error sending image to image builder: %v", err)
		app.State = model.ApplicationStateFailed
		if err := c.updateApplication(ctx, app); err != nil {
			return err
		}
		return err
	}

	return nil
}
