package httpserver

import (
	"github.com/ipaas-org/ipaas-backend/controller"
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
	"github.com/labstack/echo/v4"
)

type (
	HttpTemplateApplicationPost struct {
		Name         string           `json:"name"`
		TemplateCode string           `json:"templateCode"`
		Envs         []model.KeyValue `json:"envs,omitempty"`
	}
)

func (h *httpHandler) NewApplicationFromTemplate(c echo.Context) error {
	user, msgErr := h.ValidateAccessTokenAndGetUser(c)
	if msgErr != nil {
		return respErrorFromHttpError(c, msgErr)
	}

	ctx := c.Request().Context()

	post := new(HttpTemplateApplicationPost)
	if err := c.Bind(post); err != nil {
		return respError(c, 400, "invalid request body", "", ErrInvalidRequestBody)
	}

	if !h.controller.IsNameAvailableUserWide(ctx, post.Name, user.Code) {
		return respError(c, 400, "name taken", "name not available, there is already another service in your namespace with that name, change it please", ErrNameTaken)
	}

	template, err := h.controller.GetTemplateByCode(ctx, post.TemplateCode)
	if err != nil {
		if err == repo.ErrNotFound {
			return respError(c, 400, "template code not found", "this template code is not found, make sure the right one was selected", ErrTemplateCodeNotFound)
		}
		return respError(c, 500, "unexpected error", "", ErrUnexpected)
	}

	app, err := h.controller.CreateNewApplicationBasedOnTemplate(ctx, user.Code, post.Name, template, post.Envs)
	if err != nil {
		switch err {
		case controller.ErrMissingRequiredEnvForTemplate:
			return respError(c, 400, "missing required envs", "missing required envs for this template", ErrMissingRequiredEnvForTemplate)
		}
		return respError(c, 500, "unexpected error", "", ErrUnexpected)
	}

	resp := map[string]interface{}{
		"applicationID": app.ID.Hex(),
		"state":         app.State,
	}
	return respSuccess(c, 200, "application created successfully", resp)

}
