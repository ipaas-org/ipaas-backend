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
func (c *Controller) CreateRefreshToken(ctx context.Context, userCode string) (*model.RefreshToken, error) {
	refreshTokenDuration := time.Hour * 24 * 7

	ran, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	refreshTokenValue := ran.String()

	refreshToken := new(model.RefreshToken)
	refreshToken.Token = refreshTokenValue
	refreshToken.ExpiresAt = time.Now().Add(refreshTokenDuration)
	refreshToken.UserCode = userCode

	return refreshToken, nil
}

func (c *Controller) CreateAccessToken(ctx context.Context, userCode string) (*model.AccessToken, error) {
	token, accessTokenDuration, err := c.jwtHandler.GenerateToken(userCode)
	if err != nil {
		return nil, err
	}
	accessToken := new(model.AccessToken)
	accessToken.Token = token
	accessToken.ExpiresAt = accessTokenDuration
	return accessToken, nil
}

// todo: should add a access token model that has the expiration in it
// todo: return pointer to access token model and pointer to refresh token
func (c *Controller) GenerateTokenPair(ctx context.Context, userCode string) (*model.AccessToken, *model.RefreshToken, error) {
	accessToken, err := c.CreateAccessToken(ctx, userCode)
	if err != nil {
		return nil, nil, err
	}

	refreshToken, err := c.CreateRefreshToken(ctx, userCode)
	if err != nil {
		return nil, nil, err
	}

	_, err = c.TokenRepo.InsertOne(ctx, refreshToken)
	if err != nil {
		c.l.Errorf("error inserting refresh token: %v", err)
		return nil, nil, err
	}

	return accessToken, refreshToken, nil
}

func (c *Controller) IsRefreshTokenExpired(ctx context.Context, refreshToken string) (bool, error) {
	token, err := c.TokenRepo.FindByToken(ctx, refreshToken)
	if err != nil {
		if err == repo.ErrNotFound {
			return true, nil
		}
		return true, err
	}

	return token.ExpiresAt.Before(time.Now()), nil
}

func (c *Controller) IsAccessTokenExpired(ctx context.Context, accessToken string) (bool, error) {
	return c.jwtHandler.IsTokenExpired(accessToken)
}

// todo: should add a access token model that has the expiration in it
// todo: return pointer to access token model and pointer to refresh token
func (c *Controller) GenerateTokenPairFromRefreshToken(ctx context.Context, refreshToken string) (*model.AccessToken, *model.RefreshToken, error) {
	token, err := c.TokenRepo.FindByToken(ctx, refreshToken)
	if err != nil {
		return nil, nil, err
	}

	if token.ExpiresAt.Before(time.Now()) {
		return nil, nil, ErrTokenExpired
	}

	if _, err = c.TokenRepo.DeleteByToken(ctx, refreshToken); err != nil {
		return nil, nil, err
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
