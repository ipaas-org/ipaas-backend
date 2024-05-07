package gitProvider

import (
	"context"
	"errors"

	"github.com/ipaas-org/ipaas-backend/model"
)

const (
	ProviderGithub = "github"
)

type Provider interface {
	//*oauth functions
	GenerateLoginRedirectUri(ctx context.Context, state string) string
	GetAccessTokenFromCode(ctx context.Context, code string) (string, error)
	GetUserInfo(ctx context.Context, accessToken string) (*model.UserInfo, error)

	//*git functions
	// returns the commit of the pulled repo
	PullRepository(ctx context.Context, accessToken, username, repo, branch, path string) (string, error)
	//given a repo like username/repo returns the username and the repo name
	GetUserAndRepo(ctx context.Context, repo string) (string, string, error)
	GetUserRepos(ctx context.Context, accessToken string) ([]model.GitRepo, error)
	//get the branches of a repo and returns the default branch, all the branches or an error
	//if the repo was not found or the user does not have access to it
	GetRepoBranches(ctx context.Context, accessToken, username, repo string) (string, []string, error)
	GetLastCommitHash(ctx context.Context, accessToken, username, repo, branch string) (string, error)
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
