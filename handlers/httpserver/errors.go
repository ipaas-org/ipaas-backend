package httpserver

type HttpErrorType string

const (
	//auth errors
	ErrMissingAuthorizationHeader HttpErrorType = "missing_authorization_header"
	ErrInvalidAuthorizationHeader HttpErrorType = "invalid_authorization_header"
	ErrInvalidAccessToken         HttpErrorType = "invalid_access_token"
	ErrAccessTokenExpired         HttpErrorType = "access_token_expired"
	ErrAccessTokenNotFound        HttpErrorType = "access_token_not_found"
	ErrInvalidRefreshToken        HttpErrorType = "invalid_refresh_token"
	ErrRefreshTokenExpired        HttpErrorType = "refresh_token_expired"

	//state errors
	ErrInvalidState HttpErrorType = "invalid_state"

	//user related errors
	ErrInvalidUser                      HttpErrorType = "invalid_user"
	ErrUserNotFound                     HttpErrorType = "user_not_found"
	ErrNameAlreadyTaken                 HttpErrorType = "name_already_taken"
	ErrUnableToBuildImageInCurrentState HttpErrorType = "unable_to_build_image_in_current_state"

	// application related errors
	ErrNameTaken HttpErrorType = "name_taken"

	//http errors
	ErrInvalidRequestBody HttpErrorType = "invalid_request_body"
	ErrUnexpected         HttpErrorType = "unexpected_error"
	ErrNotImplemented     HttpErrorType = "not_implemented"
)
