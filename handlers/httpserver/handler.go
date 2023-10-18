package httpserver

import (
	"net/http"
	"time"

	"github.com/ipaas-org/ipaas-backend/controller"
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type (
	HttpTokenResponse struct {
		AccessToken           string    `json:"access_token"`
		AccessTokenExpiresIn  time.Time `json:"access_token_expires_in"`
		RefreshToken          string    `json:"refresh_token"`
		RefreshTokenExpiresIn time.Time `json:"refresh_token_expires_in"`
	}

	HttpRefreshTokensRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	httpHandler struct {
		e          *echo.Echo
		controller *controller.Controller
		l          *logrus.Logger
		Done       chan struct{}
	}
)

func NewHttpHandler(e *echo.Echo, c *controller.Controller, l *logrus.Logger) *httpHandler {
	return &httpHandler{
		e:          e,
		controller: c,
		l:          l,
		Done:       make(chan struct{}),
	}
}

// Registers only the routes and links functions
func (h *httpHandler) RegisterRoutes() {
	h.e.GET("/swagger/*", echoSwagger.WrapHandler)
	h.e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	api := h.e.Group("/api/v1")

	api.GET("/login", h.Login)
	api.GET("/oauth/callback", h.OauthCallback)
	api.POST("/token/refresh", h.RefreshTokens)
	//authenticated user routes
	authGroup := api.Group("", h.jwtHeaderCheckerMiddleware)

	authUser := authGroup.Group("/user")
	authUser.GET("/info", h.UserInfo)
	authUser.POST("/update", h.UpdateUser)

	//deployment routes
	deployment := authGroup.Group("/deployment")

	webDeployment := deployment.Group("/web")
	webDeployment.POST("/new", h.NewWebDepolyment)

	dbDeployment := deployment.Group("/db")
	dbDeployment.POST("/new", h.NewDbDeployment)
}

func (h *httpHandler) RefreshTokens(c echo.Context) error {
	ctx := c.Request().Context()

	req := new(HttpRefreshTokensRequest)
	if err := c.Bind(req); err != nil {
		h.l.Errorf("error binding body: %v", err)
		return respError(c, 400, "invalid request body", ErrInvalidRequestBody, "invalid request body")
	}

	h.l.Debugf("refresh token: %s", req.RefreshToken)
	jwt, jwtExpiresAt, refresh, refreshExpiresAt, err := h.controller.GenerateTokenPairFromRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		switch err {
		case repo.ErrNotFound:
			return respError(c, 400, "invalid refresh token", ErrInvalidRefreshToken, "invalid refresh token")
		case controller.ErrTokenExpired:
			return respError(c, 400, "expired refresh token", ErrRefreshTokenExpired, "refresh token is expired, please login again")
		}
		h.l.Errorf("error generating token pair from refresh token: %v", err)
		return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to generate token pair from refresh token")
	}

	resp := HttpTokenResponse{
		AccessToken:           jwt,
		AccessTokenExpiresIn:  jwtExpiresAt,
		RefreshToken:          refresh,
		RefreshTokenExpiresIn: refreshExpiresAt,
	}

	return respSuccess(c, 200, "successfully refreshed tokens", resp)
}

func (h *httpHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()

	user, msgErr := h.ValidateAccessTokenAndGetUser(c)
	if msgErr != nil {
		if msgErr.ErrorType == ErrUnexpected {
			h.l.Errorf("unexpected error trying to validate access token and get user: %s", msgErr.Details)
			return respErrorFromHttpError(c, msgErr)
		}
		uri := h.controller.GenerateLoginUri(ctx)
		return respSuccess(c, 200, "login uri generated", uri)
	}

	h.l.Infof("user %v already logged in, redirecting to homepage", user)
	return c.Redirect(http.StatusFound, "/api/v1/user/info")
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

	info, err := h.controller.GetUserInfoFromOauthCode(ctx, code)
	if err != nil {
		h.l.Errorf("error getting user from oauth code: %v", err)
		return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to get user from oauth code")
	}

	found := true
	user, err := h.controller.GetUserFromEmail(ctx, info.Email)
	if err != nil {
		if err == repo.ErrNotFound {
			found = false
		} else {
			h.l.Errorf("error getting user from email: %v", err)
			return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to get user from email")
		}
	}

	if !found {
		h.l.Infof("user [%s] not found, creating new user", info.Email)
		user = new(model.User)
		user.Info = info
		userCode, err := h.controller.CreateNewUserCode(ctx)
		if err != nil {
			h.l.Errorf("error creating new user code: %v", err)
			return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to create new user code")
		}
		user.Code = userCode

		networkID, err := h.controller.CreateNewNetwork(ctx, userCode)
		if err != nil {
			h.l.Errorf("error creating new network: %v", err)
			return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to create new network")
		}
		user.NetworkID = networkID

		err = h.controller.CreateUser(ctx, user)
		if err != nil {
			h.l.Errorf("error creating new user: %v", err)
			return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to create new user")
		}
	} else {
		h.l.Debugf("user %v  already exists", user)
	}

	//generate jwt and refresh token
	jwt, jwtExpiresAt, refresh, refreshExpiresAt, err := h.controller.GenerateTokenPair(ctx, user.Code)
	if err != nil {
		return respError(c, 500, "unexpected error", ErrUnexpected, "unexpected error trying to generate token pair")
	}

	resp := HttpTokenResponse{
		AccessToken:           jwt,
		AccessTokenExpiresIn:  jwtExpiresAt,
		RefreshToken:          refresh,
		RefreshTokenExpiresIn: refreshExpiresAt,
	}

	return respSuccess(c, 200, "successfully logged in", resp)
}
