package httpserver

import (
	"time"

	"github.com/ipaas-org/ipaas-backend/controller"
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/labstack/echo/v4"
)

type (
	HttpDbPost struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Version     string `json:"version"`
		Default     string `json:"default"`
		Description string `json:"description,omitempty"`
	}

	HttpAppPost struct {
		Name        string           `json:"name"`
		RepoUrl     string           `json:"repo"`
		Branch      string           `json:"branch"`
		Language    string           `json:"language"`
		Port        string           `json:"port"`
		Description string           `json:"description,omitempty"`
		Envs        []model.KeyValue `json:"envs,omitempty"`
	}
)

func (h *httpHandler) NewWebDepolyment(e echo.Context) error {
	ctx := e.Request().Context()

	accessToken, err := e.Cookie("ipaas-access-token")
	if err != nil {
		return respError(e, 400, "access token not found", ErrAccessTokenNotFound, "access token not found")
	}

	user, err := h.controller.GetUserFromAccessToken(ctx, accessToken.Value)
	if err != nil {
		return respError(e, 500, "unexpected error", ErrUnexpected, "unexpected error trying to get user from access token")
	}

	var post HttpAppPost
	err = e.Bind(&post)
	if err != nil {
		return respError(e, 400, "invalid request body", ErrInvalidRequestBody, "invalid request body")
	}
	if !h.controller.IsNameAvailableSystemWide(ctx, post.Name) {
		return respErrorf(e, 400, "name already taken", ErrNameAlreadyTaken, "name %s can not be used as it is already taken", post.Name)
	}

	var app model.Application
	app.Name = post.Name
	// app.Description = post.Description
	app.OwnerUsername = user.ID.Hex()
	if post.Port == "" {
		post.Port = "80"
	} else {
		app.Port = post.Port
	}
	app.GithubRepo = post.RepoUrl
	app.GithubBranch = post.Branch
	app.CreatedAt = time.Now()
	app.IsUpdatable = false

	h.l.Debugf("creating new service with name %s", app.Name)
	h.l.Debugf("app: %+v", app)

	err = h.controller.BuildImage(ctx, &app, user.GithubAccessToken)
	if err != nil {
		if err == controller.ErrUnableToBuildImageInCurrentState {
			return respErrorf(e, 400, "unable to build image", ErrUnableToBuildImageInCurrentState, "you can not build an image while in status %s", app.Status)
		}
		return respError(e, 500, "unexpected error", ErrUnexpected, "unexpected error trying to create service")
	}
	return respSuccessf(e, 200, "service created, in status %s", app.Status)
}
