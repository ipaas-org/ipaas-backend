package oauth

import "github.com/ipaas-org/ipaas-backend/model"

const (
	ProviderGithub = "github"
)

type Oauther interface {
	GenerateLoginRedirectUri(state string) string
	GetAccessTokenFromCode(code string) (string, error)
	GetUserInfo(accessToken string) (model.User, error)
}
