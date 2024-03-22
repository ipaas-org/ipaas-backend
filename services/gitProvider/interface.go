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
	//given a repo like username/repo returns the username and the repo name
	GetUserAndRepo(repo string) (string, string, error)
	GetUserRepos(accessToken string) ([]model.GitRepo, error)
	//get the branches of a repo and returns the default branch, all the branches or an error
	//if the repo was not found or the user does not have access to it
	GetRepoBranches(accessToken, username, repo string) (string, []string, error)
	GetLastCommitHash(accessToken, username, repo, branch string) (string, error)
	//todo
	//SetListenerToRepo() //webhook
	//RemoveListenerFromRepo()
}

var (
	ErrNotImplemented   error = errors.New("not implemented")
	ErrRateLimitReached error = errors.New("rate limit reached")
	ErrRepoNotFound     error = errors.New("repo not found")
	ErrNoCommitsFound   error = errors.New("no commits found")
)
