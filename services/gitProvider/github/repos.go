package github

import (
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/services/gitProvider"
)

func (g *GithubProvider) GetUserRepos(accessToken string) ([]model.GitRepo, error) {
	return nil, gitProvider.ErrNotImplemented
}
