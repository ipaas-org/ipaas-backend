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

	HttpTemplate struct {
		Code          string            `json:"code"`
		Name          string            `json:"name"`
		RequiredEnvs  []model.KeyValue  `json:"requiredEnvs"`
		OptionalEnvs  []model.KeyValue  `json:"optionalEnvs"`
		Description   string            `json:"description"`
		Documentation string            `json:"documentation"`
		Kind          model.ServiceKind `json:"kind"`
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

func (h *httpHandler) ListTemplates(c echo.Context) error {
	ctx := c.Request().Context()
	templates, err := h.controller.ListTemplates(ctx)
	if err != nil {
		return respError(c, 500, "unexpected error", "", ErrUnexpected)
	}
	var respTemplates []*HttpTemplate
	for _, t := range templates {
		respTemplates = append(respTemplates, &HttpTemplate{
			Code:          t.Code,
			Name:          t.Name,
			RequiredEnvs:  t.RequiredEnvs,
			OptionalEnvs:  t.OptionalEnvs,
			Description:   t.Description,
			Documentation: t.Documentation,
			Kind:          t.Kind,
		})
	}
	return respSuccess(c, 200, "templates listed successfully", respTemplates)
}

func (h *httpHandler) GetTemplate(c echo.Context) error {
	ctx := c.Request().Context()
	code := c.Param("code")
	if code == "" {
		return respError(c, 400, "invalid template code", "template code is required", ErrTemplateCodeNotFound)
	}
	template, err := h.controller.GetTemplateByCode(ctx, code)
	if err != nil {
		if err == repo.ErrNotFound {
			return respError(c, 400, "template code not found", "this template code is not found, make sure the right one was selected", ErrTemplateCodeNotFound)
		}
		return respError(c, 500, "unexpected error", "", ErrUnexpected)
	}
	resp := &HttpTemplate{
		Code:          template.Code,
		Name:          template.Name,
		RequiredEnvs:  template.RequiredEnvs,
		OptionalEnvs:  template.OptionalEnvs,
		Description:   template.Description,
		Documentation: template.Documentation,
		Kind:          template.Kind,
	}
	return respSuccess(c, 200, "template retrieved successfully", resp)
}
