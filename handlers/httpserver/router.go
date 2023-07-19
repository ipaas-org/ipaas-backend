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
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogRemoteIP: true,
		LogMethod:   true,
		LogURI:      true,
		LogStatus:   true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			l.WithFields(logrus.Fields{
				"ip":     v.RemoteIP,
				"method": v.Method,
				"uri":    v.URI,
				"status": v.Status,
			}).Info("request")
			return nil
		},
	}))
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper:      middleware.DefaultSkipper,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))

	echo.NotFoundHandler = func(c echo.Context) error {
		return respError(c, http.StatusNotFound, "invalid endpoint", fmt.Sprintf("endpoint %s is not handled, check the documentation", c.Request().URL.Path), "invalid_endpoint")
	}

	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	api := e.Group("/api/v1")

	httpHandler := NewHttpHandler(api, controller, l)
	httpHandler.RegisterRoutes()
}
