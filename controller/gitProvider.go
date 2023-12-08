package controller

import (
	"context"

	"github.com/ipaas-org/ipaas-backend/model"
)

// todo
func (c *Controller) GetAvailableGitRepos(ctx context.Context, accessToken string) ([]model.GitRepo, error) {
	// repos, err := c.gitProvider.GetAvailableRepos(ctx, accessToken)
	return nil, ErrNotImplemented
}
