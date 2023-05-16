package httpserver

import (
	"github.com/labstack/echo/v4"
)

func (h *httpHandler) jwtHeaderCheckerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	// minBearerLength := 10
	return func(c echo.Context) error {
		// authHeader := c.Request().Header.Get("Authorization")
		// if len(authHeader) < minBearerLength {
		// 	return respError(c, 401, "missing authorization header", "missing authorization header, check if the authorization header is set", "missing_authorization_header")
		// }
		// if !strings.HasPrefix(authHeader, "Bearer ") {
		// 	return respError(c, 401, "broken bearer", fmt.Sprintf("authorization header malformed, your tokens starts with %s, it needs to be \"Bearer <token>\"", authHeader[:8]), "broken_bearer")
		// }
		// accessToken := strings.TrimPrefix(authHeader, "Bearer ")

		accessToken, err := c.Cookie("ipaas-access-token")
		if err != nil {
			return respError(c, 401, "missing access token", "missing access token, you probably need to do login again", "missing_access_token")
		}
		expired, err := h.controller.IsAccessTokenExpired(c.Request().Context(), accessToken.Value)
		if err != nil {
			h.l.Errorf("unexpected error trying to check if token is expired: %v", err)
			h.l.Debugf("token: %s", c.Request().Header.Get("Authorization"))
			return respError(c, 401, "invalid token", "invalid token", "invalid_token")
		}
		if expired {
			return respError(c, 401, "token expired", "your token has expired, please login again", "token_expired")
		}

		return next(c)
	}
}
