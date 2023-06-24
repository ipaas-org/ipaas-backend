package httpserver

import (
	"net/http"
	"time"

	"github.com/ipaas-org/ipaas-backend/controller"
	"github.com/ipaas-org/ipaas-backend/repo"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

const (
	ErrUnexpected                       = "unexpected_error"
	ErrInvalidState                     = "invalid_state"
	ErrAccessTokenNotFound              = "access_token_not_found"
	ErrInvalidRequestBody               = "invalid_request_body"
	ErrNameAlreadyTaken                 = "name_already_taken"
	ErrUnableToBuildImageInCurrentState = "unable_to_build_image_in_current_state"
)

type (
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
	h.e.GET("/login", h.Login)
	h.e.GET("/oauth/callback", h.OauthCallback)

	//authenticated user routes
	authUser := h.e.Group("/user", h.jwtHeaderCheckerMiddleware)
	authUser.GET("/info", h.UserInfo)
	authUser.POST("/update", h.UpdateUser)

	//deployment routes
	deployment := h.e.Group("/deployment", h.jwtHeaderCheckerMiddleware)
	deployment.POST("/web/new", h.NewWebDepolyment, h.jwtHeaderCheckerMiddleware)
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
		return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to check access token")
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
		if err == repo.ErrNotFound {
			uri := h.controller.GenerateLoginUri(ctx)
			return respSuccess(c, 200, "login uri generated", uri)
		}
		h.l.Errorf("error getting user from access token: %v", err)
		return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to get user from access token")
	}

	h.l.Infof("user %s already logged in, redirecting to homepage", user.Email)
	// return respSuccess(c, 200, "user already logged in, it should go to /home", user)
	return c.Redirect(http.StatusFound, "/api/v1/info")
}

// TODO: check if username has changed and if the email is the same, if so update the username
// and the githubUrl (given by the oauth)
func (h *httpHandler) OauthCallback(c echo.Context) error {
	ctx := c.Request().Context()
	code := c.QueryParam("code")
	state := c.QueryParam("state")

	h.l.Debugf("found state %s and code %s", state, code)
	valid, err := h.controller.CheckState(ctx, state)
	if err != nil {
		return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to check state")
	}
	if !valid {
		return respError(c, 400, "invalid state", ErrInvalidState, "invalid state")
	}
	h.l.Debug("valid state")

	user, err := h.controller.GetUserFromOauthCode(ctx, code)
	if err != nil {
		h.l.Errorf("error getting user from oauth code: %v", err)
		return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to get user from oauth code")
	}

	found, err := h.controller.DoesUserExist(ctx, user.Email)
	if err != nil {
		h.l.Errorf("error checking if the user (%s) already exists: %v", user.Email, err)
		return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to check if user exists")
	}
	if !found {
		h.l.Debug("user not found, creating new user")
		networkID, err := h.controller.CreateNewNetwork(ctx, user.Username)
		if err != nil {
			h.l.Errorf("error creating new network: %v", err)
			return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to create new network")
		}
		user.NetworkID = networkID
		err = h.controller.CreateUser(ctx, &user)
		if err != nil {
			h.l.Errorf("error creating new user: %v", err)
			return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to create new user")
		}
	} else {
		h.l.Debugf("user %s (name=%s email=%s)  already exists", user.Username, user.FullName, user.Email)
	}

	foundUser, err := h.controller.GetUserFromEmail(ctx, user.Email)
	if err != nil {
		return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to get user from email")
	}

	//generate jwt and refresh token
	jwt, refresh, err := h.controller.GenerateTokenPair(ctx, foundUser.Email)
	if err != nil {
		return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to generate token pair")
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

	return respSuccess(c, 200, "successfully logged in")
}
