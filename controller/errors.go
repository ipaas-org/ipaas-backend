package controller

import (
	"errors"
)

var (
	//user errors
	ErrNetworkIDNotSet = errors.New("network id not set")
	ErrUserInfoNotSet  = errors.New("user info not set")
	ErrUserNotFound    = errors.New("user not found")

	//token errors
	ErrTokenExpired = errors.New("token expired")
	ErrInvalidToken = errors.New("invalid token")

	//application errors
	ErrApplicationNameNotAvailable = errors.New("name is not available")
	ErrUnsupportedApplicationKind  = errors.New("unsupported application kind")

	//templates errors
	ErrMissingRequiredEnvForTemplate = errors.New("missing required env for template")

	//image builder errors
	ErrInvalidOperationInCurrentState = errors.New("invalid operation in current state")

	//generics
	ErrNotImplemented = errors.New("not implemented")
)
