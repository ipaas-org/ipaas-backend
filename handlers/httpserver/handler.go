package httpserver

import (
	"net/http"
	"time"

	"github.com/ipaas-org/ipaas-backend/controller"
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	bearerHeaderLength = 7
)

type (
	// Here we declare a user model that will be returned by the api
	// to any unauthorized user as some informations should be visible only to admins
	HttpUnauthenticatedUser struct {
		ID        int    `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Pfp       string `json:"pfp"`
		Email     string `json:"email"`
	}
	// We are not going to declare a model for the authorized request as we will just return the model

	HttpNewUserPost struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Password  string `json:"password"`
	}

	HttpUpdateUserPost struct {
		ID        int    `json:"id,omitempty" validate:"optional"`
		FirstName string `json:"first_name,omitempty" validate:"optional"`
		LastName  string `json:"last_name,omitempty" validate:"optional"`
		Email     string `json:"email,omitempty" validate:"optional"`
		Password  string `json:"password,omitempty" validate:"optional"`
	}

	HttpLoginUserPost struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	httpHandler struct {
		e          *echo.Group
		controller *controller.Controller
		l          *logrus.Logger
	}
)

func NewHttpHandler(e *echo.Group, c *controller.Controller, l *logrus.Logger) *httpHandler {
	return &httpHandler{
		e:          e,
		controller: c,
		l:          l,
	}
}

// Registers only the routes and links functions
func (h *httpHandler) RegisterRoutes() {
	//
	h.e.GET("/login", h.Login)
	h.e.GET("/oauth/callback", h.OauthCallback)
	h.e.GET("/info", h.UserInfo, h.jwtHeaderCheckerMiddleware)

	h.e.GET("/service/new", h.NewService, h.jwtHeaderCheckerMiddleware)
	//user routes
	// h.e.GET("/:id", h.GetUnauthorizedUser)
	// h.e.GET("/all", h.GetAllUnauthorizedUsers)
	// h.e.POST("/register", h.CreateNewUser)
	// h.e.POST("/login", h.LoginUser)

	// h.e.GET("/me", h.GetUserInfo, h.jwtHeaderCheckerMiddleware)
	// h.e.GET("/update", h.UpdateUser, h.jwtHeaderCheckerMiddleware)
}

func (h *httpHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()
	//get ipaas-access-token cookie
	accessToken, err := c.Cookie("ipaas-access-token")
	if err != nil {
		uri := h.controller.GenerateLoginUri(ctx)
		return respSuccess(c, 200, "login uri generated", uri)
	}

	h.l.Debugf("found access token %s", accessToken.Value)
	//check if token is valid
	expired, err := h.controller.IsAccessTokenExpired(ctx, accessToken.Value)
	if err != nil {
		h.l.Errorf("error checking if access token is expired: %v", err)
		return respError(c, 500, "unexpected error", "unexpected error trying to check access token", "unexpected_error")
	}
	if expired {
		h.l.Debug("access token expired, generating new login uri")
		uri := h.controller.GenerateLoginUri(ctx)
		return respSuccess(c, 200, "login uri generated", uri)
	}

	//get user from token
	h.l.Infof("access token is valid, getting user from token")
	user, err := h.controller.GetUserFromAccessToken(ctx, accessToken.Value)
	if err != nil {
		h.l.Errorf("error getting user from access token: %v", err)
		return respError(c, 500, "unexpected error", "unexpected error trying to get user from access token", "unexpected_error")
	}

	h.l.Infof("user %s already logged in, redirecting to homepage", user.Email)
	// return respSuccess(c, 200, "user already logged in, it should go to /home", user)
	return c.Redirect(http.StatusFound, "/api/v1/info")
}

func (h *httpHandler) OauthCallback(c echo.Context) error {
	ctx := c.Request().Context()
	code := c.QueryParam("code")
	state := c.QueryParam("state")

	h.l.Debugf("found state %s and code %s", state, code)
	valid, err := h.controller.CheckState(ctx, state)
	if err != nil {
		return respError(c, 500, "unexpected error", "unexpected error trying to check state", "unexpected_error")
	}
	if !valid {
		return respError(c, 400, "invalid state", "invalid state", "invalid_state")
	}
	h.l.Debug("valid state")

	user, err := h.controller.GetUserFromOauthCode(ctx, code)
	if err != nil {
		h.l.Errorf("error getting user from oauth code: %v", err)
		return respError(c, 500, "unexpected error", "unexpected error trying to get user from oauth code", "unexpected_error")
	}

	found, err := h.controller.DoesUserExist(ctx, user.Email)
	if err != nil {
		h.l.Errorf("error checking if the user (%s) already exists: %v", user.Email, err)
		return respError(c, 500, "unexpected error", "unexpected error trying to check if user exists", "unexpected_error")
	}
	if !found {
		h.l.Debug("user not found, creating new user")
		err = h.controller.CreateUser(ctx, &user)
		if err != nil {
			h.l.Errorf("error creating new user: %v", err)
			return respError(c, 500, "unexpected error", "unexpected error trying to create new user", "unexpected_error")
		}
	} else {
		h.l.Debugf("user %s (name=%s email=%s)  already exists", user.Username, user.FullName, user.Email)
	}

	foundUser, err := h.controller.GetUserFromEmail(ctx, user.Email)
	if err != nil {
		return respError(c, 500, "unexpected error", "unexpected error trying to get user from email", "unexpected_error")
	}

	//generate jwt and refresh token
	jwt, refresh, err := h.controller.GenerateTokenPair(ctx, foundUser.Email)
	if err != nil {
		return respError(c, 500, "unexpected error", "unexpected error trying to generate token pair", "unexpected_error")
	}

	//set cookies
	cookie := new(http.Cookie)
	cookie.Name = "ipaas-access-token"
	cookie.Value = jwt
	cookie.Expires = time.Now().Add(15 * time.Minute)
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.SameSite = http.SameSiteStrictMode
	c.SetCookie(cookie)

	cookie = new(http.Cookie)
	cookie.Name = "ipaas-refresh-token"
	cookie.Value = refresh
	cookie.Expires = time.Now().Add(7 * 24 * time.Hour)
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = true
	cookie.SameSite = http.SameSiteStrictMode
	c.SetCookie(cookie)

	return respSuccess(c, 200, "successfully logged in", nil)
}

func (h *httpHandler) UserInfo(c echo.Context) error {
	ctx := c.Request().Context()
	accessToken, err := c.Cookie("ipaas-access-token")
	if err != nil {
		return respError(c, 400, "access token not found", "access token not found", "access_token_not_found")
	}

	user, err := h.controller.GetUserFromAccessToken(ctx, accessToken.Value)
	if err != nil {
		return respError(c, 500, "unexpected error", "unexpected error trying to get user from access token", "unexpected_error")
	}

	return respSuccess(c, 200, "user info", user)
}

func (h *httpHandler) NewService(e echo.Context) error {
	ctx := e.Request().Context()
	accessToken, err := e.Cookie("ipaas-access-token")
	if err != nil {
		return respError(e, 400, "access token not found", "access token not found", "access_token_not_found")
	}

	user, err := h.controller.GetUserFromAccessToken(ctx, accessToken.Value)
	if err != nil {
		return respError(e, 500, "unexpected error", "unexpected error trying to get user from access token", "unexpected_error")
	}

	var post model.AppPost
	err = e.Bind(&post)
	if err != nil {
		return respError(e, 400, "invalid request body", "invalid request body", "invalid_request_body")
	}

	var app model.Application
	app.ID = primitive.NewObjectID()
	app.Name = "testing"
	// app.Description = post.Description
	app.OwnerUsername = user.Email
	app.Port = "8080"
	app.GithubRepo = "vano2903/testing"
	app.GithubBranch = ""

	if !h.controller.IsNameAvailableSystemWide(ctx, app.Name) {
		return respError(e, 400, "name already taken", "name already taken", "name_already_taken")
	}

	err = h.controller.BuildImage(ctx, &app, user.GithubAccessToken)
	if err != nil {
		return respError(e, 500, "unexpected error", "unexpected error trying to create service", "unexpected_error")
	}

	return respSuccess(e, 200, "service created, in status pending")
}
