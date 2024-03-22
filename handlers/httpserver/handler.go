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
	// todo delete
	// user.DELETE("/", h.DeleteUser)

	application := authGroup.Group("/application")
	application.GET("/list/:kind", h.ListApplications)
	application.POST("/new/web", h.NewWebApplication)
	application.POST("/new/template", h.NewApplicationFromTemplate)

	//specific application routes
	application.GET("/:applicationID", h.GetApplication)
	application.PATCH("/:applicationID/update", h.UpdateApplication)
	application.DELETE("/:applicationID/delete", h.DeleteApplication)
	application.GET("/:applicationID/redeploy", h.RedeployApplication)
	application.GET("/:applicationID/status", h.GetApplicationStatus)
	application.GET("/:applicationID/rollout", h.RolloutApplication)
	application.GET("/:applicationID/logs", h.GetApplicationLogs)
	// todo, get stats and metrics
	// application.GET("/:applicationID/stats", h.GetApplicationStats)

	validate := authGroup.Group("/validate")
	validate.POST("/name", h.IsValidName)
	validate.POST("/repo", h.IsValidGitRepo)

	templates := api.Group("/templates")
	templates.GET("/list", h.ListTemplates)
	templates.GET("/:code", h.GetTemplate)

	// adminGroup := api.Group("/admin", h.jwtHeaderCheckerMiddleware, h.adminCheckerMiddleware)
	// adminUser := adminGroup.Group("/user")
	// adminUser.GET("/list", h.AdminListUsers)
	// adminUser.GET("/:userID", h.AdminGetUser)
	// adminUser.PUT("/:userID", h.AdminUpdateUser)
	// adminUser.DELETE("/:userID", h.AdminDeleteUser)
	// adminUser.DELETE("/:userID/tokens", h.AdminDeleteUserTokens)
	// adminUser.GET("/:userID/applications", h.AdminListUserApplications)

	// adminUserApplication := adminUser.Group("/:userID/application/")
	// adminUserApplication.GET("/:applicationID", h.AdminGetUserApplication)
	// adminUserApplication.PUT("/:applicationID", h.AdminUpdateUserApplication)
	// adminUserApplication.DELETE("/:applicationID", h.AdminDeleteUserApplication)
	// adminUserApplication.GET("/:applicationID/restart", h.AdminRestartUserApplication)
	// adminUserApplication.GET("/:applicationID/rollout", h.AdminRolloutUserApplication)

	// applicationAdmin := adminGroup.Group("/application")

}
