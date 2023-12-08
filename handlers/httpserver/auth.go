package httpserver

import (
	"time"

	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
	"github.com/labstack/echo/v4"
)

type (
	HttpTokenResponse struct {
		AccessToken           string          `json:"access_token"`
		AccessTokenExpiresIn  time.Time       `json:"access_token_expires_in"`
		RefreshToken          string          `json:"refresh_token"`
		RefreshTokenExpiresIn time.Time       `json:"refresh_token_expires_in"`
		StateKind             model.StateKind `json:"state_kind"`
	}

	HttpTokenKey struct {
		Key string `json:"key"`
	}
)

func (h *httpHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()
	//X-Login-Kind: api | frontend
	xLoginKind := c.Request().Header.Get("X-Login-Kind")
	h.l.Debug("login request with x-login-kind header:", xLoginKind)

	var kind model.StateKind
	switch xLoginKind {
	case string(model.StateKindAPI):
		kind = model.StateKindAPI
	case string(model.StateKindFrontend):
		kind = model.StateKindFrontend
	default:
		kind = model.StateKindAPI
	}
	h.l.Debug("login request with kind:", kind)

	user, msgErr := h.ValidateAccessTokenAndGetUser(c)
	if msgErr != nil {
		if msgErr.ErrorType == ErrUnexpected {
			h.l.Errorf("unexpected error trying to validate access token and get user: %s", msgErr.Details)
			return respErrorFromHttpError(c, msgErr)
		}
		uri := h.controller.GenerateLoginUri(ctx, kind)
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
	valid, kind, err := h.controller.CheckState(ctx, state)
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
		h.l.Debugf("user %v already exists", user)
	}

	//generate jwt and refresh token
	jwt, refresh, err := h.controller.GenerateTokenPair(ctx, user.Code)
	if err != nil {
		return respError(c, 500, "unexpected error", "unexpected error but try to login, if it doesnt work contact an admin", ErrUnexpected)
	}

	if kind == model.StateKindFrontend {
		h.l.Debugf("storing tokens temporarly and generating redirect uri")
		uri, err := h.controller.StoreTokensGenerateRedirectUri(ctx, jwt, refresh)
		if err != nil {
			return respError(c, 500, "unexpected error", "unexpected error, try to login again, if it doesnt work contanct an admin", ErrUnexpected)
		}
		h.l.Debug("redirect to: ", uri)
		return c.Redirect(307, uri)

	} else {
		resp := HttpTokenResponse{
			AccessToken:           jwt.Token,
			AccessTokenExpiresIn:  jwt.ExpiresAt,
			RefreshToken:          refresh.Token,
			RefreshTokenExpiresIn: refresh.ExpiresAt,
			StateKind:             kind,
		}
		return respSuccess(c, 200, "successfully logged in", resp)
	}
}

func (h *httpHandler) RetrieveTokens(c echo.Context) error {
	//read key from body
	//get tokens from temp storage
	//delete tokens from temp storage
	//return tokens
	ctx := c.Request().Context()

	key := new(HttpTokenKey)
	if err := c.Bind(key); err != nil {
		return respError(c, 400, "invalid request", "invalid request body", ErrInvalidRequestBody)
	}

	h.l.Debugf("retriving tokens from key: %s", key.Key)
	accessToken, refreshToken, err := h.controller.GetTokensFromKeyAndDeleteKey(ctx, key.Key)
	if err != nil {
		if err == repo.ErrNotFound {
			return respError(c, 400, "invalid key", "the used key is either expired or invalid, try to login again", ErrInvalidKey)
		}
		h.l.Errorf("error getting tokens from key: %v", err)
		return respError(c, 500, "unexpected error", "unexpected error", ErrUnexpected)
	}

	resp := HttpTokenResponse{
		AccessToken:           accessToken.Token,
		AccessTokenExpiresIn:  accessToken.ExpiresAt,
		RefreshToken:          refreshToken.Token,
		RefreshTokenExpiresIn: refreshToken.ExpiresAt,
		StateKind:             model.StateKindFrontend,
	}
	return respSuccess(c, 200, "tokens retrived successfully", resp)
}
