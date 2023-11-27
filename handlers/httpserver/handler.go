package httpserver

import (
	"time"

	"github.com/ipaas-org/ipaas-backend/controller"
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

	application := authGroup.Group("/application")
	application.GET("/list/:kind", h.ListApplications)
	application.GET("/status/:applicationID", h.GetApplicationStatus)
	application.GET("/:applicationID", h.GetApplication)
	// application.GET("/logs/:applicationID", h.GetApplicationLogs)
	application.DELETE("/delete/:applicationID", h.DeleteApplication)
	// application.PUT("/update/:applicationID", h.UpdateApplication)
	application.POST("/new/web", h.NewWebApplication)
	application.POST("/new/template", h.NewApplicationFromTemplate)

	// validation := authGroup.Group("/validation")
	// validation.POST("/name", h.IsNameAvailable)
	// validation.POST("/repo", h.IsValidGitRepo)

	// templates := authGroup.Group("/templates")
	// templates.GET("/list", h.ListTemplates)
	// templates.GET("/:templateID", h.GetTemplate)
	// //deployment routes
	// deployment := authGroup.Group("/deployment")

	// webDeployment := deployment.Group("/web")
	// webDeployment.POST("/new", h.NewWebDepolyment)

	// dbDeployment := deployment.Group("/db")
	// dbDeployment.POST("/new", h.NewDbDeployment)
}
