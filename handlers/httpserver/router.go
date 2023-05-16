package httpserver

import (
	"fmt"
	"net/http"

	"github.com/ipaas-org/ipaas-backend/config"
	"github.com/ipaas-org/ipaas-backend/controller"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// Swagger spec:
//
//	@title			IPAAS API
//	@version		0.1.0
//	@description	api for the endpoints
//	@contact.url	https://github.com/ipaas-org/ipaas-backend
//	@host			localhost:8082
//	@BasePath		/api/v1
func InitRouter(e *echo.Echo, l *logrus.Logger, controller *controller.Controller, conf *config.Config) {
	// e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	echo.NotFoundHandler = func(c echo.Context) error {
		// render your 404 page
		return respError(c, http.StatusNotFound, "invalid endpoint", fmt.Sprintf("endpoint %s is not handled", c.Request().URL.Path), "invalid_endpoint")
	}

	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	api := e.Group("/api/v1")
	//user routes

	httpHandler := NewHttpHandler(api, controller, l)
	httpHandler.RegisterRoutes()

	// container := api.Group("/container")
}
