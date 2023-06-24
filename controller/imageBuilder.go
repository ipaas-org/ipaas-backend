package controller

import (
	"context"
	"fmt"

	"github.com/ipaas-org/ipaas-backend/model"
)

var (
	ErrUnableToBuildImageInCurrentState = fmt.Errorf("unable to build image in current state")
)

// TODO: IMPORTANTE l'user id quando si crea l'immagine non può essere la mail, meglio usare l'id dell'utente o il suo username
// TODO: non accettare richieste di build image di un applicazione se l'applicazione è già in status pending|updating
func (c *Controller) BuildImage(ctx context.Context, app *model.Application, providerToken string) error {
	if app.Status == StatusPending || app.Status == StatusBuilding || app.Status == StatusUpdating {
		return ErrUnableToBuildImageInCurrentState
	}

	request := model.BuildRequest{
		UUID:      app.ID.Hex(),
		Token:     providerToken,
		UserID:    app.OwnerUsername,
		Type:      "repo",
		Connector: "github",
		Repo:      app.GithubRepo,
		Branch:    app.GithubBranch,
	}

	if err := c.imageBuilder.BuildImage(request); err != nil {
		c.l.Errorf("error building image: %v", err)
		app.Status = StatusFailed
		if err := c.UpdateApplicationState(ctx, app); err != nil {
			return fmt.Errorf("c.updateApplicationState: %w", err)
		}
		return fmt.Errorf("c.imageBuilder.BuildImage: %w", err)
	}

	app.Status = StatusBuilding
	if err := c.UpdateApplicationState(ctx, app); err != nil {
		return fmt.Errorf("c.updateApplicationState: %w", err)
	}

	return nil
}
