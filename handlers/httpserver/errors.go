package httpserver

const (
	//auth errors
	ErrMissingAuthorizationHeader = "missing_authorization_header"
	ErrInvalidAuthorizationHeader = "invalid_authorization_header"
	ErrInvalidAccessToken         = "invalid_access_token"
	ErrAccessTokenExpired         = "access_token_expired"
	ErrAccessTokenNotFound        = "access_token_not_found"
	ErrInvalidRefreshToken        = "invalid_refresh_token"
	ErrRefreshTokenExpired        = "refresh_token_expired"

	//state errors
	ErrInvalidState = "invalid_state"

	//user related errors
	ErrInvalidUser                      = "invalid_user"
	ErrUserNotFound                     = "user_not_found"
	ErrNameAlreadyTaken                 = "name_already_taken"
	ErrUnableToBuildImageInCurrentState = "unable_to_build_image_in_current_state"

	//http errors
	ErrInvalidRequestBody = "invalid_request_body"
	ErrUnexpected         = "unexpected_error"
)
