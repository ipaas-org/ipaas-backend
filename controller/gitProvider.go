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

// repo is name/repo
func (c *Controller) ValidateGitRepo(ctx context.Context, user *model.User, repo string) (string, []string, error) {
	// return c.gitProvider.ValidateRepo(ctx, accessToken, username, repo)
	username, repo, err := c.gitProvider.GetUserAndRepo(ctx, repo)
	if err != nil {
		return "", nil, err
	}

	defaultBranch, branches, err := c.gitProvider.GetRepoBranches(ctx, user.Info.GithubAccessToken, username, repo)
	if err != nil {
		c.l.Errorf("error getting branches from git provider: %v", err)
		return "", nil, err
	}
	return defaultBranch, branches, nil
}

func (c *Controller) GetLastCommitHash(ctx context.Context, user *model.User, repo string, branch string) (string, error) {
	// return c.gitProvider.GetLastCommitHash(ctx, accessToken, username, repo, branch)
	username, repo, err := c.gitProvider.GetUserAndRepo(ctx, repo)
	if err != nil {
		return "", err
	}

	commitHash, err := c.gitProvider.GetLastCommitHash(ctx, user.Info.GithubAccessToken, username, repo, branch)
	if err != nil {
		c.l.Errorf("error getting last commit hash from git provider: %v", err)
		return "", err
	}
	return commitHash, nil
}
