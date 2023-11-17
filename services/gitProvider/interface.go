package gitProvider

import (
	"errors"

	"github.com/ipaas-org/ipaas-backend/model"
)

const (
	ProviderGithub = "github"
)

type Provider interface {
	//oauth functions
	GenerateLoginRedirectUri(state string) string
	GetAccessTokenFromCode(code string) (string, error)
	GetUserInfo(accessToken string) (*model.UserInfo, error)

	//git functions
	GetUserRepos(accessToken string) ([]model.GitRepo, error)
	//todo
	//SetListenerToRepo() //webhook
	//RemoveListenerFromRepo()
}

var (
	ErrNotImplemented error = errors.New("not implemented")
)
