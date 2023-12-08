package httpserver

import (
	"github.com/ipaas-org/ipaas-backend/controller"
	"github.com/ipaas-org/ipaas-backend/repo"
	"github.com/labstack/echo/v4"
)

type (
	HttpRefreshTokensRequest struct {
		RefreshToken string `json:"refresh_token"`
	}
)

func (h *httpHandler) RefreshTokens(c echo.Context) error {
	ctx := c.Request().Context()

	req := new(HttpRefreshTokensRequest)
	if err := c.Bind(req); err != nil {
		h.l.Errorf("error binding body: %v", err)
		return respError(c, 400, "invalid request body", "", ErrInvalidRequestBody)
	}

	h.l.Debugf("refresh token: %s", req.RefreshToken)
	jwt, refresh, err := h.controller.GenerateTokenPairFromRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		switch err {
		case repo.ErrNotFound:
			return respError(c, 400, "invalid refresh token", "", ErrInvalidRefreshToken)
		case controller.ErrTokenExpired:
			return respError(c, 400, "expired refresh token", "refresh token is expired, please login again", ErrRefreshTokenExpired)
		}
		h.l.Errorf("error generating token pair from refresh token: %v", err)
		return respError(c, 500, "unexpected error", "unexpected error trying to generate token pair from refresh token", ErrUnexpected)
	}

	resp := HttpTokenResponse{
		AccessToken:           jwt.Token,
		AccessTokenExpiresIn:  jwt.ExpiresAt,
		RefreshToken:          refresh.Token,
		RefreshTokenExpiresIn: refresh.ExpiresAt,
	}

	return respSuccess(c, 200, "successfully refreshed tokens", resp)
}
