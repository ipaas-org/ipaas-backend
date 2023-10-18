package controller

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ipaas-org/ipaas-backend/model"
	"github.com/ipaas-org/ipaas-backend/repo"
)

// todo: function needs to return pointer
// todo: function should return only refresh token and not the expiration as it's inside the structure
func (c *Controller) CreateRefreshToken(ctx context.Context, userEmail string) (model.RefreshToken, time.Time, error) {
	refreshTokenDuration := time.Hour * 24 * 7

	ran, err := uuid.NewRandom()
	if err != nil {
		return model.RefreshToken{}, time.Time{}, err
	}

	refreshTokenValue := ran.String()
	var refreshToken model.RefreshToken
	refreshToken.Token = refreshTokenValue
	refreshToken.Expiration = time.Now().Add(refreshTokenDuration)
	refreshToken.UserCode = userEmail

	return refreshToken, refreshToken.Expiration, nil
}

// todo: should add a access token model that has the expiration in it
// todo: return pointer to access token model and pointer to refresh token
func (c *Controller) GenerateTokenPair(ctx context.Context, userCode string) (string, time.Time, string, time.Time, error) {
	accessToken, accessTokenDuration, err := c.jwtHandler.GenerateToken(userCode)
	if err != nil {
		return "", time.Time{}, "", time.Time{}, err
	}

	refreshToken, refreshTokenExpiresAt, err := c.CreateRefreshToken(ctx, userCode)
	if err != nil {
		return "", time.Time{}, "", time.Time{}, err
	}

	_, err = c.TokenRepo.InsertOne(ctx, &refreshToken)
	if err != nil {
		return "", time.Time{}, "", time.Time{}, err
	}

	return accessToken, accessTokenDuration, refreshToken.Token, refreshTokenExpiresAt, err
}

func (c *Controller) IsRefreshTokenExpired(ctx context.Context, refreshToken string) (bool, error) {
	token, err := c.TokenRepo.FindByToken(ctx, refreshToken)
	if err != nil {
		if err == repo.ErrNotFound {
			return true, nil
		}
		return true, err
	}

	return token.Expiration.Before(time.Now()), nil
}

func (c *Controller) IsAccessTokenExpired(ctx context.Context, accessToken string) (bool, error) {
	return c.jwtHandler.IsTokenExpired(accessToken)
}

// todo: should add a access token model that has the expiration in it
// todo: return pointer to access token model and pointer to refresh token
func (c *Controller) GenerateTokenPairFromRefreshToken(ctx context.Context, refreshToken string) (string, time.Time, string, time.Time, error) {
	token, err := c.TokenRepo.FindByToken(ctx, refreshToken)
	if err != nil {
		return "", time.Time{}, "", time.Time{}, err
	}

	if token.Expiration.Before(time.Now()) {
		return "", time.Time{}, "", time.Time{}, ErrTokenExpired
	}

	if _, err = c.TokenRepo.DeleteByToken(ctx, refreshToken); err != nil {
		return "", time.Time{}, "", time.Time{}, err
	}
	return c.GenerateTokenPair(ctx, token.UserCode)
}

func (c *Controller) ValidateAccessTokenAndGetUser(ctx context.Context, accessToken string) (*model.User, error) {
	claims, err := c.jwtHandler.ValidateToken(accessToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if c.jwtHandler.IsTokenExpiredFromClaims(claims) {
		return nil, ErrTokenExpired
	}

	user, err := c.UserRepo.FindByCode(ctx, claims.UserCode)
	if err != nil {
		if err == repo.ErrNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

//TODO: Implement revoke token
// the refreshToken can easly be deleted but for the access token
// we either store it in a blacklist or we store all the access tokens in the database.
// if we store all the tokens then we can also implement a "logout from all devices" feature
// while if we do it with the blacklist we need to not allow
// all the tokens issued before the "logout from all devices" to be valid (which is not a bad thing)
