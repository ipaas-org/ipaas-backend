package controller

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
)

// send request to image builder, to build a specific commit use the commit hash, leave blank
// to build the latest commit
func (c *Controller) BuildImage(ctx context.Context, app *model.Application, commit, providerToken string) error {
	if app.State == model.ApplicationStateBuilding ||
		app.State == model.ApplicationStateStarting {
		return ErrInvalidOperationInCurrentState
	}

	request := model.BuildRequest{
		ApplicationID: app.ID.Hex(),
		PullInfo: &model.PullInfoRequest{
			Token:     providerToken,
			UserID:    app.Owner,
			Connector: "github",
			Repo:      app.GithubRepo,
			Branch:    app.GithubBranch,
			Commit:    commit,
		},
		BuildPlan: app.BuildPlan,
	}

	c.l.Debugf("sending to rmq: %+v ", request)
	if err := c.imageBuilder.BuildImage(ctx, request); err != nil {
		c.l.Errorf("error sending image to image builder: %v", err)
		app.State = model.ApplicationStateFailed
		if err := c.updateApplication(ctx, app); err != nil {
			c.l.Errorf("error updating application state: %v", err)
			return err
		}
		return err
	}
	return nil
}
