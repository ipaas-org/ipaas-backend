package httpserver

import (
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/labstack/echo/v4"
)

type (
	HttpWebApplicationPost struct {
		Name        string           `json:"name"`
		Repo        string           `json:"repo"`
		Branch      string           `json:"branch"`
		Language    string           `json:"language"`
		Port        string           `json:"port"`
		Description string           `json:"description,omitempty"`
		Envs        []model.KeyValue `json:"envs,omitempty"`
	}
)

// todo
func (h *httpHandler) IsApplicationNameAvailable(c echo.Context) error {
	return respError(c, 501, "not implemented", "", ErrNotImplemented)
}

// todo
func (h *httpHandler) IsValidGitRepo(c echo.Context) error {
	return respError(c, 501, "not implemented", "", ErrNotImplemented)
}

func (h *httpHandler) NewWebApplication(c echo.Context) error {
	user, msgErr := h.ValidateAccessTokenAndGetUser(c)
	if msgErr != nil {
		return respErrorFromHttpError(c, msgErr)
	}

	ctx := c.Request().Context()

	post := new(HttpWebApplicationPost)
	if err := c.Bind(post); err != nil {
		return respError(c, 400, "invalid request body", "", ErrInvalidRequestBody)
	}

	if !h.controller.IsNameAvailableSystemWide(ctx, post.Name) {
		return respError(c, 400, "name taken", "name not available as it's already been taken", ErrNameTaken)
	}

	app, err := h.controller.CreateNewWebApplication(ctx, user.Code, user.Info.GithubAccessToken, post.Name, post.Repo, post.Branch, post.Port, post.Envs)
	if err != nil {
		return respError(c, 500, "unexpected error", "", ErrUnexpected)
	}
	return respSuccessf(c, 200, "application created, the current state is [%s]", string(app.State))
}
