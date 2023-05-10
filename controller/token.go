package controller

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
)

var (
	ErrTokenExpired = errors.New("token expired")
)

func (c *Controller) TokenGenerateTokenPair(ctx context.Context, userEmail string) (string, string, error) {
	accessToken, err := c.jwtHandler.GenerateToken(userEmail)
	if err != nil {
		return "", "", err
	}

	ran, err := uuid.NewRandom()
	if err != nil {
		return "", "", err
	}
	refreshTokenValue := ran.String()

	var refreshToken model.RefreshToken
	refreshToken.Token = refreshTokenValue
	refreshToken.Expiration = time.Now().Add(time.Hour * 24 * 7)
	refreshToken.UserEmail = userEmail

	_, err = c.tokenRepo.Insert(ctx, &refreshToken)
	return accessToken, refreshTokenValue, err
}

func (c *Controller) TokenIsRefreshTokenExpired(ctx context.Context, refreshToken string) (bool, error) {
	token, err := c.tokenRepo.FindByToken(ctx, refreshToken)
	if err != nil {
		if err == repo.ErrNotFound {
			return true, nil
		}
		return true, err
	}

	return token.Expiration.Before(time.Now()), nil
}

func (c *Controller) TokenGetUserFromAccessToken(ctx context.Context, accessToken string) (*model.User, error) {
	claims, err := c.jwtHandler.ValidateToken(accessToken)
	if err != nil {
		return nil, err
	}

	if c.jwtHandler.IsTokenExpiredFromClaims(claims) {
		return nil, ErrTokenExpired
	}

	return c.userRepo.FindByEmail(ctx, claims.UserEmail)
}

func (c *Controller) TokenGenerateTokenPairFromRefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	token, err := c.tokenRepo.FindByToken(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}

	if token.Expiration.Before(time.Now()) {
		return "", "", ErrTokenExpired
	}

	if _, err = c.tokenRepo.DeleteByToken(ctx, refreshToken); err != nil {
		return "", "", err
	}
	return c.TokenGenerateTokenPair(ctx, token.UserEmail)
}

//TODO: Implement revoke token
// the refreshToken can easly be deleted but for the access token
// we either store it in a blacklist or we store all the access tokens in the database.
// if we store all the tokens then we can also implement a "logout from all devices" feature
// while if we do it with the blacklist we need to not allow
// all the tokens issued before the "logout from all devices" to be valid (which is not a bad thing)
