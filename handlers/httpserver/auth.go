package httpserver

import (
	"github.com/ipaas-org/ipaas-backend/repo"
	"github.com/labstack/echo/v4"
)

func (h *httpHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()

	user, msgErr := h.ValidateAccessTokenAndGetUser(c)
	if msgErr != nil {
		if msgErr.ErrorType == ErrUnexpected {
			h.l.Errorf("unexpected error trying to validate access token and get user: %s", msgErr.Details)
			return respErrorFromHttpError(c, msgErr)
		}
		uri := h.controller.GenerateLoginUri(ctx)
		return respSuccess(c, 200, "login uri generated", uri)
	}

	h.l.Infof("user %v already logged in, redirecting to homepage", user)
	return c.Redirect(404, "/api/v1/user/info")
}

// TODO: check if username has changed and if the email is the same, if so update the username
// and the githubUrl (given by the oauth)
func (h *httpHandler) OauthCallback(c echo.Context) error {
	ctx := c.Request().Context()
	code := c.QueryParam("code")
	state := c.QueryParam("state")

	// h.l.Debugf("found state %s and code %s", state, code)
	valid, err := h.controller.CheckState(ctx, state)
	if err != nil {
		h.l.Errorf("error checking state: %v", err)
		return respError(c, 500, "unexpected error", "", ErrUnexpected)
	}
	if !valid {
		return respError(c, 400, "invalid state", "", ErrInvalidState)
	}

	info, err := h.controller.GetUserInfoFromOauthCode(ctx, code)
	if err != nil {
		h.l.Errorf("error getting user from oauth code: %v", err)
		return respError(c, 500, "unexpected error", "", ErrUnexpected)
	}

	found := true
	user, err := h.controller.GetUserFromEmail(ctx, info.Email)
	if err != nil {
		if err == repo.ErrNotFound {
			found = false
		} else {
			h.l.Errorf("error getting user from email: %v", err)
			return respError(c, 500, "unexpected error", "", ErrUnexpected)
		}
	}

	if !found {
		h.l.Infof("user [%s] not found, creating new user", info.Email)
		user, err = h.controller.CreateUser(ctx, info, "", "")
		if err != nil {
			h.l.Errorf("error creating new user: %v", err)
			return respError(c, 500, "unexpected error", "", ErrUnexpected)
		}
	} else {
		h.l.Debugf("user %v  already exists", user)
	}

	//generate jwt and refresh token
	jwt, refresh, err := h.controller.GenerateTokenPair(ctx, user.Code)
	if err != nil {
		return respError(c, 500, "unexpected error", "unexpected error but try to login, if it doesnt work contact an admin", ErrUnexpected)
	}

	resp := HttpTokenResponse{
		AccessToken:           jwt.Token,
		AccessTokenExpiresIn:  jwt.ExpiresAt,
		RefreshToken:          refresh.Token,
		RefreshTokenExpiresIn: refresh.ExpiresAt,
	}

	return respSuccess(c, 200, "successfully logged in", resp)
}
