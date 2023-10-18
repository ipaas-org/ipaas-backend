package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/ipaas-org/ipaas-backend/config"
	"github.com/ipaas-org/ipaas-backend/controller"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

// Swagger spec:
//
//	@title			IPAAS API
//	@version		0.1.0
//	@description	api for the endpoints
//	@contact.url	https://github.com/ipaas-org/ipaas-backend
//	@host			localhost:8082
//	@BasePath		/api/v1
func InitRouter(e *echo.Echo, l *logrus.Logger, controller *controller.Controller, conf *config.Config) *httpHandler {
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogRemoteIP: true,
		LogMethod:   true,
		LogURI:      true,
		LogStatus:   true,
		LogLatency:  true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			l.Infof("ip=%q method=%q uri=%q status=%d latency=%q",
				v.RemoteIP,
				v.Method,
				v.URI,
				v.Status,
				v.Latency)
			return nil
		},
	}))
	// e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper:      middleware.DefaultSkipper,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))

	echo.NotFoundHandler = func(c echo.Context) error {
		return respError(c, http.StatusNotFound, "invalid endpoint", fmt.Sprintf("endpoint %s is not handled, check the documentation", c.Request().URL.Path), "invalid_endpoint")
	}

	httpHandler := NewHttpHandler(e, controller, l)
	httpHandler.RegisterRoutes()
	return httpHandler
}

func StartRouter(ctx context.Context, h *httpHandler, conf *config.Config, ID int, RoutineMonitor chan int) {
	defer func() {
		if r := recover(); r != nil {
			h.l.Errorf("router panic, recovering: \nerror: %v\n\nstack: %s", r, string(debug.Stack()))
		}
		if ctx.Err() == nil {
			RoutineMonitor <- ID
		} else {
			h.l.Info("API not restarting, context was canceled")
		}
	}()
	h.l.Info(">>> STARTING HTTP SERVER >>>")

	s := http.Server{
		Addr:    "0.0.0.0:" + conf.HTTP.Port,
		Handler: h.e,
	}

	h.l.Infof(">>> STARTING API ON: %s >>>", s.Addr)
	err := s.ListenAndServe()
	if err != nil {
		if ctx.Err() != nil {
			h.l.Info(">>> STOPPING SERVER >>>")
			h.Done <- struct{}{}
		} else if err == http.ErrServerClosed {
			h.l.Info(">>> SERVER CLOSED >>>")
		} else {
			h.l.Errorf(">>> ERROR STARTING SERVER: %v", err.Error())
		}
		return
	}
}
