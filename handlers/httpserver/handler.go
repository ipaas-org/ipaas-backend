package httpserver

import (
	"github.com/ipaas-org/ipaas-backend/controller"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type (
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
	api.POST("/retriveTokens", h.RetrieveTokens)
	api.GET("/oauth/callback", h.OauthCallback)
	api.POST("/token/refresh", h.RefreshTokens)

	authGroup := api.Group("", h.jwtHeaderCheckerMiddleware)
	//authenticated user routes
	user := authGroup.Group("/user")
	user.GET("/info", h.UserInfo)
	user.POST("/update", h.UpdateUser)

	application := authGroup.Group("/application")
	application.GET("/list/:kind", h.ListApplications)
	application.GET("/status/:applicationID", h.GetApplicationStatus)
	application.GET("/:applicationID", h.GetApplication)
	// application.GET("/logs/:applicationID", h.GetApplicationLogs)
	application.DELETE("/delete/:applicationID", h.DeleteApplication)
	// application.PUT("/update/:applicationID", h.UpdateApplication)
	application.POST("/new/web", h.NewWebApplication)
	application.POST("/new/template", h.NewApplicationFromTemplate)

	validate := authGroup.Group("/validate")
	validate.POST("/name", h.IsValidName)
	validate.POST("/repo", h.IsValidGitRepo)

	templates := api.Group("/templates")
	templates.GET("/list", h.ListTemplates)
	templates.GET("/:code", h.GetTemplate)
	// //deployment routes
	// deployment := authGroup.Group("/deployment")

	// webDeployment := deployment.Group("/web")
	// webDeployment.POST("/new", h.NewWebDepolyment)

	// dbDeployment := deployment.Group("/db")
	// dbDeployment.POST("/new", h.NewDbDeployment)
}
