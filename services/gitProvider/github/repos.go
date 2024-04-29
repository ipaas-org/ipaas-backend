package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/services/gitProvider"
	"github.com/tidwall/gjson"
)

type (
	GithubCommit struct {
		SHA    string               `json:"sha"`
		Commit GithubCommitInternal `json:"commit"`
	}

	GithubCommitInternal struct {
		Message string `json:"message"`
	}
)

const (
	baseUrlMetadata = "https://api.github.com/repos/%s/%s"
	baseUrlCommits  = "https://api.github.com/repos/%s/%s/commits?sha=%s"
	baseUrlBranches = "https://api.github.com/repos/%s/%s/branches"
	baseUrlTag      = "https://api.github.com/repos/%s/%s/tags"
	baseUrlRelease  = "https://api.github.com/repos/%s/%s/releases"
)

func (g *GithubProvider) GetUserRepos(ctx context.Context, accessToken string) ([]model.GitRepo, error) {
	return nil, gitProvider.ErrNotImplemented
}

func (g *GithubProvider) GetRepoBranches(ctx context.Context, accessToken, username, repo string) (string, []string, error) {
	// return "", nil, gitProvider.ErrNotImplemented
	wg := sync.WaitGroup{}
	wg.Add(2)
	var defaultBranch string
	var branches []string
	var err error
	go func() {
		defer wg.Done()
		defaultBranch, err = g.getDefaultBranch(ctx, accessToken, username, repo)
	}()
	go func() {
		defer wg.Done()
		branches, err = g.getBranches(ctx, username, repo, accessToken)
	}()
	wg.Wait()
	return defaultBranch, branches, err
}

func (g *GithubProvider) getDefaultBranch(ctx context.Context, accessToken, username, repo string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(baseUrlMetadata, username, repo), nil)
	if err != nil {
		return "", err
	}

	// request.Header.Set("User-Agent", g.userAgent)
	request.Header.Set("Authorization", "token "+accessToken)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	jsonBody := string(body)

	if resp.StatusCode != 200 {
		if resp.StatusCode == 403 {
			// g.l.Errorf("githubConnector.getBranchAndDescription: github api rate limit exceeded: %v", jsonBody)
			return "", gitProvider.ErrRateLimitReached
		}
		if resp.StatusCode == 404 {
			// g.l.Errorf("githubConnector.getBranchAndDescription: repo not found: %v", jsonBody)
			return "", gitProvider.ErrRepoNotFound
		}
		// g.l.Errorf("githubConnector.getBranchAndDescription: error finding release info for %s/%s [%s]: %v", username, repo, resp.Status, jsonBody)
		return "", fmt.Errorf("error finding general info for %s/%s [%s]: %v", username, repo, resp.Status, jsonBody)
	}

	return gjson.Get(jsonBody, "default_branch").String(), nil
}

func (g *GithubProvider) getBranches(ctx context.Context, username, repo, token string) ([]string, error) {
	//get the branches
	request, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(baseUrlBranches, username, repo), nil)
	if err != nil {
		return nil, err
	}

	// request.Header.Set("User-Agent", g.userAgent)
	request.Header.Set("Authorization", "token "+token)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	jsonBody := string(body)

	if resp.StatusCode != 200 {
		if resp.StatusCode == 403 {
			return nil, gitProvider.ErrRateLimitReached
		}
		if resp.StatusCode == 404 {
			return nil, gitProvider.ErrRepoNotFound
		}
		return nil, fmt.Errorf("error finding branches info for %s/%s [%s]: %v", username, repo, resp.Status, jsonBody)
	}

	branchesRes := gjson.Get(jsonBody, "@this.#.name").Array()
	branches := make([]string, len(branchesRes))
	for i, r := range branchesRes {
		branches[i] = r.String()
	}
	return branches, nil
}

// GetUserAndRepo get the username of the creator and the repository's name given a GitHub repository url
func (g *GithubProvider) GetUserAndRepo(ctx context.Context, url string) (string, string, error) {
	url = strings.TrimSuffix(url, ".git")
	split := strings.Split(url, "/")

	return split[len(split)-2], split[len(split)-1], nil
}

func (g *GithubProvider) GetLastCommitHash(ctx context.Context, accessToken, username, repo, branch string) (string, error) {
	url := fmt.Sprintf(baseUrlCommits, username, repo, branch)
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	request.Header.Set("Authorization", "token "+accessToken)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	jsonBody := string(body)

	if resp.StatusCode != 200 {
		if resp.StatusCode == 403 {
			return "", gitProvider.ErrRateLimitReached
		}
		if resp.StatusCode == 404 {
			return "", gitProvider.ErrRepoNotFound
		}
		return "", fmt.Errorf("error getting last commit info for %s/%s [%s]: %v", username, repo, resp.Status, jsonBody)
	}

	var RepoCommits []GithubCommit
	err = json.Unmarshal(body, &RepoCommits)
	if err != nil {
		return "", err
	}

	if len(RepoCommits) == 0 {
		return "", gitProvider.ErrNoCommitsFound
	}

	return RepoCommits[0].SHA, nil
}
