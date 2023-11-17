package httpserver

import (
	"runtime/debug"

	"github.com/labstack/echo/v4"
)

func (h *httpHandler) jwtHeaderCheckerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	// minBearerLength := 10
	return func(c echo.Context) error {
		defer func() {
			if r := recover(); r != nil {
				h.l.Errorf("router panic, recovering: \nerror: %v\n\nstack: %s", r, string(debug.Stack()))
			}
		}()

		accessToken, httperr := h.GetAccessToken(c)
		if httperr != nil {
			return respErrorFromHttpError(c, httperr)
		}

		expired, err := h.controller.IsAccessTokenExpired(c.Request().Context(), accessToken)
		if err != nil {
			h.l.Errorf("unexpected error trying to check if token is expired, it's probably an invalid token: %v", err)
			return respError(c, 401, "invalid access token", "invalid access token, please login again", ErrInvalidAccessToken)
		}

		if expired {
			return respError(c, 401, "expired access token", "access token is expired, refresh the tokens or login again if you dont have a refresh token", ErrAccessTokenExpired)
		}

		return next(c)
	}
}
