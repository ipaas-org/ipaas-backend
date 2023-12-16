package httpserver

import (
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/services/gitProvider"
	"github.com/labstack/echo/v4"
)

type (
	HttpNameValidationgRequest struct {
		Kind model.ServiceKind `json:"kind"`
		Name string            `json:"name"`
	}

	HttpNameValidatingResponse struct {
		Valid bool `json:"valid"`
	}

	HttpGitRepoValidationRequest struct {
		Repo string `json:"repo"`
	}
	HttpGitRepoVlidationResponse struct {
		Valid         bool     `json:"valid"`
		DefaultBranch string   `json:"defaultBranch"`
		Branches      []string `json:"branches"`
	}
)

func (h *httpHandler) IsValidName(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(HttpNameValidationgRequest)
	if err := c.Bind(req); err != nil {
		return respError(c, 400, "invalid body", "request body does not match the expected format", ErrInvalidRequestBody)
	}
	if req.Name == "" {
		return respError(c, 400, "invalid body", "name cannot be empty", ErrInvalidRequestBody)
	}
	if req.Kind == "" {
		return respError(c, 400, "invalid body", "kind cannot be empty", ErrInvalidRequestBody)
	}
	switch req.Kind {
	case model.ApplicationKindWeb:
		if !h.controller.IsNameAvailableSystemWide(ctx, req.Name) {
			return respSuccess(c, 200, "name is not available", &HttpNameValidatingResponse{Valid: false})
		}
		return respSuccess(c, 200, "name is available", &HttpNameValidatingResponse{Valid: true})
	case model.ApplicationKindStorage, model.ApplicationKindManagment:
		user, err := h.ValidateAccessTokenAndGetUser(c)
		if err != nil {
			h.l.Errorf("error validating user when validating name: %v", err.Message)
			return respErrorFromHttpError(c, err)
		}
		if !h.controller.IsNameAvailableUserWide(ctx, req.Name, user.Code) {
			return respSuccess(c, 200, "name is not available", &HttpNameValidatingResponse{Valid: false})
		}
		return respSuccess(c, 200, "name is available", &HttpNameValidatingResponse{Valid: true})
	default:
		return respError(c, 400, "invalid body", "kind is not valid", ErrInvalidRequestBody)
	}
}

func (h *httpHandler) IsValidGitRepo(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(HttpGitRepoValidationRequest)
	if err := c.Bind(req); err != nil {
		return respError(c, 400, "invalid body", "request body does not match the expected format", ErrInvalidRequestBody)
	}
	user, httpErr := h.ValidateAccessTokenAndGetUser(c)
	if httpErr != nil {
		h.l.Errorf("error validating user when validating git repo: %v", httpErr.Message)
		return respErrorFromHttpError(c, httpErr)
	}

	defaultBranch, branches, err := h.controller.ValidateGitRepo(ctx, user, req.Repo)
	if err != nil {
		h.l.Errorf("error validating git repo: %v", err)
		switch err {
		case gitProvider.ErrRateLimitReached:
			return respError(c, 500, "git provider rate limit reached", "looks like you have reached the rate limit on your access token, try again in a few minutes", ErrRateLimitReached)
		case gitProvider.ErrRepoNotFound:
			return respSuccess(c, 200, "repo is not valid or not found", &HttpGitRepoVlidationResponse{Valid: false})
		default:
			return respError(c, 500, "unexpected error", "unexpected error trying to validate git repo", ErrUnexpected)
		}
	}
	return respSuccess(c, 200, "git repo is valid", &HttpGitRepoVlidationResponse{Valid: true, DefaultBranch: defaultBranch, Branches: branches})
}

//todo
// func (h *httpHandler) GetAvailableGitRepos(c echo.Context) error {
// 	//get user from context
// 	user, err := h.ValidateAccessTokenAndGetUser(c)
// 	if err != nil {
// 		h.l.Errorf("error validating user when getting available git repos: %v", err.Message)
// 		return respErrorFromHttpError(c, err)
// 	}
// 	ctx := c.Request().Context()

// 	repos, err := h.controller.GetAvailableGitRepos(ctx, user.Info.GithubAccessToken)
// 	if err != nil {
// 		switch err {
// 		//...
// 		}
// 		h.l.Errorf("unexpected error trying to get available git repos: %v", err)
// 		return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to get available git repos")
// 	}

// 	resp := new(AvailableGitReposResponse)
// 	resp.Repos = repos
// 	return respSuccess(c, 200, "successfully got available git repos", resp)
// }
