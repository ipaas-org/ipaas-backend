package parking

// package httpserver

// import (
// 	"time"

// 	"github.com/ipaas-org/ipaas-backend/controller"
// 	"github.com/ipaas-org/ipaas-backend/model"
// 	"github.com/labstack/echo/v4"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// )

// type (
// 	HttpDbPost struct {
// 		Name        string `json:"name"`
// 		Type        string `json:"type"`
// 		Version     string `json:"version"`
// 		Default     string `json:"default"`
// 		Description string `json:"description,omitempty"`
// 	}

// 	HttpAppPost struct {
// 		Name        string           `json:"name"`
// 		RepoUrl     string           `json:"repo"`
// 		Port        string           `json:"port"`
// 		Branch      string           `json:"branch,omitempty"`
// 		Description string           `json:"description,omitempty"`
// 		Envs        []model.KeyValue `json:"envs,omitempty"`
// 	}
// )

// func (h *httpHandler) NewWebDepolyment(c echo.Context) error {
// 	user, msgErr := h.ValidateAccessTokenAndGetUser(c)
// 	if msgErr != nil {
// 		return respErrorFromHttpError(c, msgErr)
// 	}

// 	ctx := c.Request().Context()

// 	post := new(HttpAppPost)
// 	err := c.Bind(post)
// 	if err != nil {
// 		h.l.Errorf("error binding body: %v", err)
// 		return respError(c, 400, "invalid request body", ErrInvalidRequestBody, "invalid request body")
// 	}
// 	if !h.controller.IsNameAvailableSystemWide(ctx, post.Name) {
// 		return respErrorf(c, 400, "name already taken", ErrNameAlreadyTaken, "name %s can not be used as it is already taken", post.Name)
// 	}
// 	if post.Port == "" {
// 		post.Port = "80"
// 	}

// 	app := new(model.Application)
// 	app.ID = primitive.NewObjectID()
// 	app.Name = post.Name
// 	app.Owner = user.Code
// 	app.Kind = model.ApplicationKindWeb
// 	app.State = model.ApplicationStatePending
// 	// app.Description = post.Description
// 	app.GithubRepo = post.RepoUrl
// 	app.GithubBranch = post.Branch
// 	app.CreatedAt = time.Now()
// 	app.PortToMap = post.Port

// 	h.l.Debugf("creating new service with name %s", app.Name)
// 	h.l.Debugf("app: %+v", app)

// 	if err := h.controller.CreateNewDeployment(ctx, app); err != nil {
// 		switch err {
// 		case controller.ErrNotImplemented:
// 			return respErrorf(c, 501, "not implemented", "not_implemented", "the %s service kind is not implemented yet", app.Kind.String())

// 		}
// 	}
// 	// err = h.controller.BuildImage(ctx, app, user.Info.GithubAccessToken)
// 	// if err != nil {
// 	// 	if err == controller.ErrUnableToBuildImageInCurrentState {
// 	// 		return respErrorf(c, 400, "unable to build image", ErrUnableToBuildImageInCurrentState, "you can not build an image while your application is in status %s", app.State.String())
// 	// 	}
// 	// 	return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to create service")
// 	// }
// 	return respSuccessf(c, 200, "service created, in state %s", app.State.String())
// }
