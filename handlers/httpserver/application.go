package httpserver

import (
	"fmt"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	HttpWebApplicationPost struct {
		Name   string `json:"name"`
		Repo   string `json:"repo"`
		Branch string `json:"branch"`
		// Language    string           `json:"language"`
		Port        string           `json:"port"`
		Description string           `json:"description,omitempty"`
		Envs        []model.KeyValue `json:"envs,omitempty"`
	}
)

func (h *httpHandler) NewWebApplication(c echo.Context) error {
	user, msgErr := h.ValidateAccessTokenAndGetUser(c)
	if msgErr != nil {
		return respErrorFromHttpError(c, msgErr)
	}

	ctx := c.Request().Context()

	post := new(HttpWebApplicationPost)
	if err := c.Bind(post); err != nil {
		h.l.Debugf("error binding request body: %v", err)
		return respError(c, 400, "invalid request body", "", ErrInvalidRequestBody)
	}
	if !h.controller.IsNameAvailableSystemWide(ctx, post.Name) {
		return respError(c, 400, "name taken", "name not available as it's already been taken", ErrNameTaken)
	}

	app, err := h.controller.CreateNewWebApplication(ctx, user.Code, user.Info.GithubAccessToken, post.Name, post.Repo, post.Branch, post.Port, post.Envs)
	if err != nil {
		//TODO: handle error
		return respError(c, 500, "unexpected error", "", ErrUnexpected)
	}

	resp := map[string]interface{}{
		"applicationID": app.ID.Hex(),
		"state":         app.State,
	}
	return respSuccess(c, 200, "application created successfully", resp)
}

func (h *httpHandler) GetApplicationStatus(c echo.Context) error {
	user, msgErr := h.ValidateAccessTokenAndGetUser(c)
	if msgErr != nil {
		return respErrorFromHttpError(c, msgErr)
	}

	ctx := c.Request().Context()

	applicationIDHex := c.Param("applicationID")
	if applicationIDHex == "" {
		return respError(c, 400, "invalid application id", "applicationID is required", ErrInvalidApplicationID)
	}

	applicationID, err := primitive.ObjectIDFromHex(applicationIDHex)
	if err != nil {
		return respError(c, 400, "invalid application id", "applicationID is invalid", ErrInvalidApplicationID)
	}

	app, err := h.controller.GetApplicationByID(ctx, applicationID)
	if err != nil {
		return respError(c, 500, "unexpected error", "", ErrUnexpected)
	}

	if app.Owner != user.Code {
		return respError(c, 403, "inexisting application id", fmt.Sprintf("the application with id=%s does not exists", applicationID), ErrInexistingApplication)
	}

	resp := map[string]interface{}{
		"applicationID": app.ID.Hex(),
		"state":         app.State,
	}
	return respSuccess(c, 200, "retreived state succesfully", resp)
}

func (h *httpHandler) ListApplications(c echo.Context) error {
	kind := c.Param("kind")
	ctx := c.Request().Context()

	user, httpErr := h.ValidateAccessTokenAndGetUser(c)
	if httpErr != nil {
		return respErrorFromHttpError(c, httpErr)
	}

	var (
		apps []*model.Application
		err  error
		msg  string
	)

	if kind == "" || kind == "all" {
		apps, err = h.controller.GetAllUserApplications(ctx, user.Code)
		msg = "list of all the applications"
	} else {
		apps, err = h.controller.GetApplicationByKind(ctx, user.Code, model.ApplicationKind(kind))
		msg = fmt.Sprintf("list of all the applications of kind [%s]", kind)
	}

	if err != nil {
		switch err {
		case repo.ErrNotFound:
			return respError(c, 404, "not found", "no applications were found, you need to create something first", ErrNotFound)
		default:
			return respError(c, 500, "unexpected error", "", ErrUnexpected)
		}
	}
	return respSuccess(c, 200, msg, apps)
}

func (h *httpHandler) GetApplication(c echo.Context) error {
	user, msgErr := h.ValidateAccessTokenAndGetUser(c)
	if msgErr != nil {
		return respErrorFromHttpError(c, msgErr)
	}

	ctx := c.Request().Context()

	applicationIDHex := c.Param("applicationID")
	if applicationIDHex == "" {
		return respError(c, 400, "invalid application id", "applicationID is required", ErrInvalidApplicationID)
	}

	applicationID, err := primitive.ObjectIDFromHex(applicationIDHex)
	if err != nil {
		return respError(c, 400, "invalid application id", "applicationID is invalid", ErrInvalidApplicationID)
	}

	app, err := h.controller.GetApplicationByID(ctx, applicationID)
	if err != nil {
		if err == repo.ErrNotFound {
			return respError(c, 404, "inexisting applcation id", fmt.Sprintf("the application with id=%s does not exists", applicationID.Hex()), ErrInexistingApplication)
		}
		return respError(c, 500, "unexpected error", "", ErrUnexpected)
	}

	if app.Owner != user.Code {
		return respError(c, 404, "inexisting application id", fmt.Sprintf("the application with id=%s does not exists", applicationID.Hex()), ErrInexistingApplication)
	}

	return respSuccess(c, 200, "application retreived successfully", app)
}

func (h *httpHandler) DeleteApplication(c echo.Context) error {
	user, msgErr := h.ValidateAccessTokenAndGetUser(c)
	if msgErr != nil {
		return respErrorFromHttpError(c, msgErr)
	}

	ctx := c.Request().Context()
	applicationIDHex := c.Param("applicationID")
	if applicationIDHex == "" {
		return respError(c, 400, "invalid application id", "applicationID is required", ErrInvalidApplicationID)
	}

	applicationID, err := primitive.ObjectIDFromHex(applicationIDHex)
	if err != nil {
		return respError(c, 400, "invalid application id", "applicationID is invalid", ErrInvalidApplicationID)
	}

	app, err := h.controller.GetApplicationByID(ctx, applicationID)
	if err != nil {
		return respError(c, 500, "unexpected error", "", ErrUnexpected)
	}

	if user.Code != app.Owner {
		return respError(c, 403, "forbidden", "you are not allowed to delete this application", ErrForbidden)
	}
	//todo
	return respError(c, 501, "not implemented", "", ErrNotImplemented)
	// if err := h.controller.DeleteApplication(ctx, app); err != nil {
	// 	switch err {
	// 	case controller.ErrInvalidOperationInCurrentState:
	// 		return respError(c, 400, "invalid operation in current state", "the application is in a state that does not allow this operation", ErrInvalidOperationInCurrentState)
	// 	default:
	// 		return respError(c, 500, "unexpected error", "", ErrUnexpected)
	// 	}
	// }
	return respSuccess(c, 200, "application deleted successfully", nil)
}

// func (h *httpHandler) UpdateApplication(c echo.Context) error {
// 	return respError(c, 501, "not implemented", "", ErrNotImplemented)
// }
