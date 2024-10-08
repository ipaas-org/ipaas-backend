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
	ErrInvalidKey                 HttpErrorType = "invalid_key"

	//state errors
	ErrInvalidState HttpErrorType = "invalid_state"

	//user related errors
	ErrInvalidUser      HttpErrorType = "invalid_user"
	ErrUserNotFound     HttpErrorType = "user_not_found"
	ErrNameAlreadyTaken HttpErrorType = "name_already_taken"

	// application related errors
	ErrNameTaken                       HttpErrorType = "name_taken"
	ErrInvalidApplicationID            HttpErrorType = "invalid_application_id"
	ErrInexistingApplication           HttpErrorType = "inexisting_application"
	ErrInvalidOperationInCurrentState  HttpErrorType = "invalid_operation_in_current_state"
	ErrInvalidOperationWithCurrentKind HttpErrorType = "invalid_operation_with_current_kind"
	ErrVersionUpToDate                 HttpErrorType = "version_up_to_date"
	ErrInvalidApplicationState         HttpErrorType = "invalid_application_state"
	ErrBranchNotFound                  HttpErrorType = "branch_not_found"
	ErrRootNotFound                    HttpErrorType = "root_not_found"
	ErrInvalidBuildPlan                HttpErrorType = "invalid_build_plan"
	ErrInvalidBuilder                  HttpErrorType = "invalid_builder"
	ErrInvalidDockerfilePath           HttpErrorType = "invalid_dockerfile_path"
	ErrInvalidPhaseCommand             HttpErrorType = "invalid_phase_command"

	//template related errors
	ErrTemplateCodeNotFound          HttpErrorType = "template_code_not_found"
	ErrMissingRequiredEnvForTemplate HttpErrorType = "missing_required_env_for_template"

	//http errors
	ErrInvalidRequestBody HttpErrorType = "invalid_request_body"
	ErrUnexpected         HttpErrorType = "unexpected_error"
	ErrNotImplemented     HttpErrorType = "not_implemented"
	ErrNotFound           HttpErrorType = "not_found"
	ErrForbidden          HttpErrorType = "forbidden"

	//git providers errors
	ErrRateLimitReached HttpErrorType = "rate_limit_reached"

	//log errors
	ErrInvalidXLastLogNano HttpErrorType = "invalid_x_last_log_nano"

	//update errors
	ErrInvalidOperation HttpErrorType = "invalid_operation"
)
