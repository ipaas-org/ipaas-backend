package httpserver

import (
	"net/http"
	"time"

	"github.com/ipaas-org/ipaas-backend/controller"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type (
	HttpTokenResponse struct {
		AccessToken           string    `json:"access_token"`
		AccessTokenExpiresIn  time.Time `json:"access_token_expires_in"`
		RefreshToken          string    `json:"refresh_token"`
		RefreshTokenExpiresIn time.Time `json:"refresh_token_expires_in"`
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
	h.e.GET("/login", h.Login)
	h.e.GET("/oauth/callback", h.OauthCallback)

	//authenticated user routes
	authGroup := h.e.Group("", h.jwtHeaderCheckerMiddleware)

	authUser := authGroup.Group("/user")
	authUser.GET("/info", h.UserInfo)
	authUser.POST("/update", h.UpdateUser)

	//deployment routes
	deployment := authGroup.Group("/deployment")
	deployment.POST("/web/new", h.NewWebDepolyment)
}

func (h *httpHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()

	user, msgErr := h.ValidateAccessTokenAndGetUser(c)
	if msgErr != nil {
		if msgErr.ErrorType == ErrInvalidAccessToken ||
			msgErr.ErrorType == ErrAccessTokenExpired ||
			msgErr.ErrorType == ErrUserNotFound {
			uri := h.controller.GenerateLoginUri(ctx)
			return respSuccess(c, 200, "login uri generated", uri)
		} else {
			h.l.Errorf("unexpected error trying to validate access token and get user: %s", msgErr.Details)
			return respErrorFromHttpError(c, msgErr)
		}
	}

	h.l.Infof("user %s already logged in, redirecting to homepage", user.Email)
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

	user, errMsg := h.controller.GetUserFromOauthCode(ctx, code)
	if errMsg != nil {
		h.l.Errorf("error getting user from oauth code: %v", errMsg)
		return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to get user from oauth code")
	}

	found, errMsg := h.controller.DoesUserExist(ctx, user.Email)
	if errMsg != nil {
		h.l.Errorf("error checking if the user (%s) already exists: %v", user.Email, errMsg)
		return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to check if user exists")
	}
	if !found {
		h.l.Debug("user not found, creating new user")
		networkID, errMsg := h.controller.CreateNewNetwork(ctx, user.Username)
		if errMsg != nil {
			h.l.Errorf("error creating new network: %v", errMsg)
			return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to create new network")
		}
		user.NetworkID = networkID
		errMsg = h.controller.CreateUser(ctx, &user)
		if errMsg != nil {
			h.l.Errorf("error creating new user: %v", errMsg)
			return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to create new user")
		}
	} else {
		h.l.Debugf("user %s (name=%s email=%s)  already exists", user.Username, user.FullName, user.Email)
	}

	foundUser, err := h.controller.GetUserFromEmail(ctx, user.Email)
	if errMsg != nil {
		return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to get user from email")
	}

	//generate jwt and refresh token
	jwt, refresh, errMsg := h.controller.GenerateTokenPair(ctx, foundUser.Email)
	if errMsg != nil {
		return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to generate token pair")
	}

	resp := HttpTokenResponse{
		AccessToken:           jwt,
		AccessTokenExpiresIn:  time.Now().Add(15 * time.Minute),
		RefreshToken:          refresh,
		RefreshTokenExpiresIn: time.Now().Add(7 * 24 * time.Hour),
	}

	return respSuccess(c, 200, "successfully logged in", resp)
}
