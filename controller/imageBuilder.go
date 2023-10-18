package controller

import (
	"context"
	"fmt"

	"github.com/ipaas-org/ipaas-backend/model"
)

// TODO: IMPORTANTE l'user id quando si crea l'immagine non può essere la mail, meglio usare l'id dell'utente o il suo username
// TODO: non accettare richieste di build image di un applicazione se l'applicazione è già in status pending|updating
func (c *Controller) BuildImage(ctx context.Context, app *model.Application, providerToken string) error {
	if app.State == StatePending || app.State == StateBuilding || app.State == StateUpdating {
		return ErrUnableToBuildImageInCurrentState
	}

	request := model.BuildRequest{
		UUID:      app.ID.Hex(),
		Token:     providerToken,
		UserID:    app.Owner,
		Type:      "repo",
		Connector: "github",
		Repo:      app.GithubRepo,
		Branch:    app.GithubBranch,
	}

	if err := c.imageBuilder.BuildImage(request); err != nil {
		c.l.Errorf("error building image: %v", err)
		app.State = StateFailed
		if err := c.UpdateApplicationState(ctx, app); err != nil {
			return fmt.Errorf("c.updateApplicationState: %w", err)
		}
		return fmt.Errorf("c.imageBuilder.BuildImage: %w", err)
	}

	app.State = StateBuilding
	if err := c.UpdateApplicationState(ctx, app); err != nil {
		return fmt.Errorf("c.updateApplicationState: %w", err)
	}

	return nil
}
