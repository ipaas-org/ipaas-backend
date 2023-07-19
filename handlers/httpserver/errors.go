package httpserver

const (
	//auth errors
	ErrMissingAuthorizationHeader = "missing authorization header"
	ErrInvalidAuthorizationHeader = "invalid authorization header"
	ErrInvalidAccessToken         = "invalid access token"
	ErrAccessTokenExpired         = "access token expired"
	ErrAccessTokenNotFound        = "access_token_not_found"

	//state errors
	ErrInvalidState = "invalid_state"

	//user related errors
	ErrInvalidUser                      = "invalid user"
	ErrUserNotFound                     = "user_not_found"
	ErrNameAlreadyTaken                 = "name_already_taken"
	ErrUnableToBuildImageInCurrentState = "unable_to_build_image_in_current_state"

	//http errors
	ErrInvalidRequestBody = "invalid_request_body"
	ErrUnexpected         = "unexpected_error"
)
