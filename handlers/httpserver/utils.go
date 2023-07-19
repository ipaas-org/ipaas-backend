package httpserver

import (
	"strings"

	"github.com/ipaas-org/ipaas-backend/controller"
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/labstack/echo/v4"
)

func (h *httpHandler) GetAccessToken(c echo.Context) (string, *HttpError) {
	httperr := new(HttpError)
	accessToken := c.Request().Header.Get("Authorization")
	if accessToken == "" {
		httperr.Code = 401
		httperr.Message = "no authorization header provided"
		httperr.ErrorType = ErrMissingAuthorizationHeader
		httperr.Details = "no authorization header provided, add authorization header with the bearer token"
		return "", httperr
	}

	if !strings.HasPrefix(accessToken, "Bearer ") {
		httperr.Code = 401
		httperr.Message = "invalid authorization header"
		httperr.ErrorType = ErrInvalidAuthorizationHeader
		httperr.Details = "invalid authorization header, add authorization header with the bearer token"
		return "", httperr
	}

	accessToken = strings.TrimPrefix(accessToken, "Bearer ")
	return accessToken, nil
}

func (h *httpHandler) ValidateAccessTokenAndGetUser(c echo.Context) (*model.User, *HttpError) {
	accessToken, httperr := h.GetAccessToken(c)
	if httperr != nil {
		return nil, httperr
	}

	httperr = new(HttpError)
	ctx := c.Request().Context()

	user, err := h.controller.ValidateAccessTokenAndGetUser(ctx, accessToken)
	if err != nil {
		switch err {
		case controller.ErrInvalidToken:
			httperr.Code = 401
			httperr.Message = "invalid access token"
			httperr.ErrorType = ErrInvalidAccessToken
			httperr.Details = "invalid access token, please login again"
			return nil, httperr
		case controller.ErrTokenExpired:
			httperr.Code = 401
			httperr.Message = "access token expired"
			httperr.ErrorType = ErrAccessTokenExpired
			httperr.Details = "access token expired, please login again"
			return nil, httperr
		case controller.ErrUserNotFound:
			httperr.Code = 401
			httperr.Message = "user not found"
			httperr.ErrorType = ErrUserNotFound
			httperr.Details = "user not found, the user might have been deleted, if this is an error, please contact the customer support"
			return nil, httperr
		default:
			httperr.Code = 500
			httperr.Message = "unexpected error"
			httperr.ErrorType = ErrUnexpected
			httperr.Details = "unexpected error, please contact the customer support"
			h.l.Errorf("unexpected error when validating token and getting user in http server: %v", err)
			return nil, httperr
		}
	}
	return user, nil
}
