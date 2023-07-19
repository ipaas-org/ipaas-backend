package controller

import (
	"errors"
)

var (
	//user errors
	ErrNetworkIDNotSet = errors.New("network id not set")
	ErrUserNotFound    = errors.New("user not found")

	//token errors
	ErrTokenExpired = errors.New("token expired")
	ErrInvalidToken = errors.New("invalid token")

	//application errors
	ErrApplicationNameNotAvailable = errors.New("name is not available")

	//image builder
	ErrUnableToBuildImageInCurrentState = errors.New("unable to build image in current state")
)
