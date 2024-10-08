package httpserver

import (
	"fmt"

	"github.com/ipaas-org/ipaas-backend/controller"
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	HttpRequestNewWebApplication struct {
		Name   string `json:"name"`
		Repo   string `json:"repo"`
		Branch string `json:"branch"`
		// Language    string           `json:"language"`
		Port        string           `json:"port"`
		Description string           `json:"description,omitempty"`
		Envs        []model.KeyValue `json:"envs,omitempty"`

		RootDirectory string `json:"rootDirectory"`
	}

	HttpRequestApplicationGeneralUpdate struct {
		Name string           `json:"name,omitempty"`
		Port string           `json:"port,omitempty"`
		Envs []model.KeyValue `json:"envs,omitempty"`
		// todo: add start command, as it doesnt require a new build
	}

	HttpRequestApplicationBuildUpdate struct {
		BuildConfig model.BuildConfig `json:"buildConfig"`
	}
)

func (h *httpHandler) NewWebApplication(c echo.Context) error {
	user, msgErr := h.ValidateAccessTokenAndGetUser(c)
	if msgErr != nil {
		return respErrorFromHttpError(c, msgErr)
	}

	ctx := c.Request().Context()

	post := new(HttpRequestNewWebApplication)
	if err := c.Bind(post); err != nil {
		h.l.Debugf("error binding request body: %v", err)
		return respError(c, 400, "invalid request body", "", ErrInvalidRequestBody)
	}
	if !h.controller.IsNameAvailableSystemWide(ctx, post.Name) {
		return respError(c, 400, "name taken", "name not available as it's already been taken", ErrNameTaken)
	}

	if post.RootDirectory == "" {
		post.RootDirectory = "/"
	}

	app, err := h.controller.CreateNewWebApplication(ctx, user.Code, user.Info.GithubAccessToken, post.Name, post.Repo, post.Branch, post.Port, post.Envs, post.RootDirectory)
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
		return respError(c, 404, "inexisting application id", fmt.Sprintf("the application with id=%s does not exists", applicationID), ErrInexistingApplication)
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
		return respError(c, 404, "forbidden", "you are not allowed to delete this application", ErrForbidden)
	}
	if err := h.controller.DeleteApplication(ctx, app, user); err != nil {
		switch err {
		case controller.ErrInvalidOperationInCurrentState:
			return respError(c, 400, "invalid operation in current state", "the application is in a state that does not allow this operation", ErrInvalidOperationInCurrentState)
		default:
			return respError(c, 500, "unexpected error", "", ErrUnexpected)
		}
	}
	return respSuccess(c, 200, "application deleted successfully", nil)
}

func (h *httpHandler) UpdateApplicationGeneral(c echo.Context) error {
	var patch HttpRequestApplicationGeneralUpdate
	if err := c.Bind(&patch); err != nil {
		return respError(c, 400, "invalid request body", "", ErrInvalidRequestBody)
	}

	if patch.Name == "" && patch.Port == "" && patch.Envs == nil {
		return respError(c, 400, "invalid request body", "at least one of the fields name, port or envs is required", ErrInvalidRequestBody)
	}

	user, app, err := h.GetUserAndApplication(c)
	if user == nil || app == nil {
		return err
	}

	if user.Code != app.Owner {
		return respError(c, 404, "inexisting applcation id", fmt.Sprintf("the application with id=%s does not exists", app.ID.Hex()), ErrInexistingApplication)
	}

	ctx := c.Request().Context()

	if err := h.controller.UpdateApplicationGeneral(ctx, app, user, patch.Name, patch.Port, patch.Envs); err != nil {
		h.l.Errorf("error updating application: %v", err)
		switch err {
		case controller.ErrInvalidOperationInCurrentState:
			return respError(c, 501, "unable to update name", "the application does not support updating the name at the moment", ErrNotImplemented)
		case controller.ErrInvalidPort:
			return respError(c, 400, "invalid port", fmt.Sprintf("the provided port %q is not a valid port, it needs to be an integer and be between 0 and 65535", patch.Port), ErrInvalidRequestBody)
		case controller.ErrInvalidEnv:
			return respError(c, 400, "invalid env", "the provided envs are invalid, they need to be a list of key value pairs", ErrInvalidRequestBody)
		case controller.ErrNoChanges:
			return respSuccess(c, 200, "no changes", nil)
		default:
			return respError(c, 500, "unexpected error", "", ErrUnexpected)
		}
	}
	return respSuccess(c, 200, "application updated successfully", nil)
}

func (h *httpHandler) UpdateApplicationBuild(c echo.Context) error {
	var patch HttpRequestApplicationBuildUpdate
	if err := c.Bind(&patch); err != nil {
		return respError(c, 400, "invalid request body", "", ErrInvalidRequestBody)
	}

	user, app, err := h.GetUserAndApplication(c)
	if user == nil || app == nil {
		return err
	}

	if user.Code != app.Owner {
		return respError(c, 404, "inexisting applcation id", fmt.Sprintf("the application with id=%s does not exists", app.ID.Hex()), ErrInexistingApplication)
	}

	ctx := c.Request().Context()
	if err := h.controller.UpdateApplicationBuild(ctx, app, user, &patch.BuildConfig); err != nil {
		h.l.Errorf("error updating application build: %v", err)
		switch err {
		case controller.ErrInvalidOperationWithCurrentKind:
			return respError(c, 400, "invalid operation with current kind", "the application kind does not support this operation", ErrInvalidOperationWithCurrentKind)
		case controller.ErrInvalidOperationInCurrentState:
			return respError(c, 400, "invalid operation in current state", fmt.Sprintf("the application is in %q state, this operation is not allowed in that state", app.State), ErrInvalidOperationInCurrentState)
		case controller.ErrInvalidBuildPlan:
			return respError(c, 400, "invalid build plan", "the provided build plan is invalid", ErrInvalidRequestBody)
		case controller.ErrInvalidBuilder:
			return respError(c, 400, "invalid builder", "the provided builder is invalid, currently only supported builders are 'nixpacks' and 'docker'", ErrInvalidRequestBody)
		case controller.ErrInvalidDockerfilePath:
			return respError(c, 400, "invalid dockerfile path", "the provided dockerfile path is invalid, it needs to be a valid path to a file in the repository", ErrInvalidRequestBody)
		case controller.ErrInvalidPhaseCommand:
			return respError(c, 400, "invalid phase command", "the provided phase command is invalid, it needs to be a valid command", ErrInvalidRequestBody)
		default:
			return respError(c, 500, "unexpected error", "", ErrUnexpected)
		}
	}
	return respSuccess(c, 200, "application build plan is being updated", nil)
}

func (h *httpHandler) RedeployApplication(c echo.Context) error {
	user, app, err := h.GetUserAndApplication(c)
	if user == nil || app == nil {
		return err
	}

	if user.Code != app.Owner {
		return respError(c, 404, "inexisting applcation id", fmt.Sprintf("the application with id=%s does not exists", app.ID.Hex()), ErrInexistingApplication)
	}

	ctx := c.Request().Context()

	if err := h.controller.RedeployApplication(ctx, user, app); err != nil {
		return respError(c, 500, "unexpeted error", "", ErrUnexpected)
	}

	return respSuccess(c, 200, "application is restarting")
}

func (h *httpHandler) RolloutApplication(c echo.Context) error {
	user, app, err := h.GetUserAndApplication(c)
	if user == nil || app == nil {
		return err
	}

	if user.Code != app.Owner {
		return respError(c, 404, "inexisting applcation id", fmt.Sprintf("the application with id=%s does not exists", app.ID.Hex()), ErrInexistingApplication)
	}

	ctx := c.Request().Context()

	if err := h.controller.RolloutApplication(ctx, user, app); err != nil {
		switch err {
		case controller.ErrInvalidOperationWithCurrentKind:
			return respError(c, 400, "invalid operation with current kind", "the application kind does not support this operation", ErrInvalidOperationWithCurrentKind)
		case controller.ErrInvalidOperationInCurrentState:
			return respError(c, 400, "invalid operation in current state", fmt.Sprintf("the application is in %q state, this operation is not allowed in that state", app.State), ErrInvalidOperationInCurrentState)
		case controller.ErrLastVersionAlreadyDeployed:
			return respError(c, 400, "last version already up to date", "the last version of the application is already up to date", ErrVersionUpToDate)
		default:
			return respError(c, 500, "unexpected error", "", ErrUnexpected)
		}
	}
	resp := map[string]interface{}{
		"applicationID": app.ID.Hex(),
		"state":         app.State,
	}
	return respSuccess(c, 200, "application is rolling out", resp)
}
