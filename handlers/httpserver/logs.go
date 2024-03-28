package httpserver

import (
	"fmt"
	"time"

	"github.com/ipaas-org/ipaas-backend/repo"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (h *httpHandler) GetApplicationLogs(c echo.Context) error {
	user, msgErr := h.ValidateAccessTokenAndGetUser(c)
	if msgErr != nil {
		return respErrorFromHttpError(c, msgErr)
	}

	ctx := c.Request().Context()

	applicationIDHex := c.Param("applicationID")
	if applicationIDHex == "" {
		return respError(c, 400, "invalid application id", "applicationID is required", ErrInvalidApplicationID)
	}

	applicationID, err := primitive.ObjectIDFromHex(applicationIDHex)
	if err != nil {
		return respError(c, 400, "invalid application id", "applicationID is invalid", ErrInvalidApplicationID)
	}

	app, err := h.controller.GetApplicationByID(ctx, applicationID)
	if err != nil {
		if err == repo.ErrNotFound {
			return respError(c, 404, "inexisting applcation id", fmt.Sprintf("the application with id=%s does not exists", applicationID.Hex()), ErrInexistingApplication)
		}
		return respError(c, 500, "unexpected error", "", ErrUnexpected)
	}

	if app.Owner != user.Code {
		return respError(c, 404, "inexisting application id", fmt.Sprintf("the application with id=%s does not exists", applicationID.Hex()), ErrInexistingApplication)
	}

	//get from and to from query params
	from := c.QueryParam("from")
	to := c.QueryParam("to")
	if from == "" {
		from = "now-1h"
	}
	if to == "" {
		to = "now"
	}

	lastLogNano := c.Request().Header.Get("X-Last-Log-Nano")
	if lastLogNano != "" {
		lastLogNanoTime, err := time.Parse(time.RFC3339Nano, lastLogNano)
		if err != nil {
			return respError(c, 400, "invalid X-Last-Log-Nano", fmt.Sprintf("unable to parse %q as a valid X-Last-Log-Nano, the format must follow the specific of RFC3339 Nano", lastLogNano), ErrInvalidXLastLogNano)
		}

		logs, err := h.controller.GetLogs(ctx, user.Namespace, app.Service.Deployment.Name, from, to)
		if err != nil {
			return respError(c, 500, "unexpected error", "", ErrUnexpected)
		}

		if lastLogNanoTime.Equal(logs.LastTimestamp) {
			return respSuccess(c, 200, "no new logs", nil)
		}
		return respSuccess(c, 200, "logs retreived successfully", logs)
	}

	logs, err := h.controller.GetLogs(ctx, user.Namespace, app.Service.Deployment.Name, from, to)
	if err != nil {
		return respError(c, 500, "unexpected error", "", ErrUnexpected)
	}

	return respSuccess(c, 200, "logs retreived successfully", logs)
}
