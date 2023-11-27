package controller

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
)

// TODO: IMPORTANTE l'user id quando si crea l'immagine non può essere la mail, meglio usare l'id dell'utente o il suo username
// TODO: non accettare richieste di build image di un applicazione se l'applicazione è già in status building o starting
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
		if err := c.UpdateApplication(ctx, app); err != nil {
			return err
		}
		return err
	}

	return nil
}
