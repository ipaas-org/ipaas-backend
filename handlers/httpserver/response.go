package httpserver

import (
	"github.com/labstack/echo/v4"
)

const (
// baseErrDocsUrl = "https://example.com/docs/errors/"
)

// ConnectionID string `json:"connection_id,omitemtpy"`
type HttpError struct {
	Code        int           `json:"code" example:"400"`
	IsError     bool          `json:"isError" example:"true"`
	Message     string        `json:"message" example:"Bad Request"`
	Details     string        `json:"details,omitempty" example:"Bad Request With More Info"`
	Instance    string        `json:"instance,omitempty" example:"/api/v1/users/1"`
	ErrorType   HttpErrorType `json:"errorType,omitempty" example:"invalid_id"`
	ErrorDocUrl string        `json:"errorDocUrl,omitempty" example:"https://example.com/docs/errors/invalid_id"`
}

func respError(c echo.Context, code int, message, details string, errType HttpErrorType) error {
	h := HttpError{
		Instance:  c.Request().RequestURI,
		IsError:   true,
		Code:      code,
		Message:   message,
		Details:   details,
		ErrorType: errType,
		// ErrorDocUrl: baseErrDocsUrl + errType, //could be a map somwhere, just an example for now
	}

	return c.JSON(code, h)
}

// func respErrorf(c echo.Context, code int, message, details string, errType HttpErrorType, args ...string) error {
// 	h := HttpError{
// 		Instance:  c.Request().RequestURI,
// 		IsError:   true,
// 		Code:      code,
// 		Message:   message,
// 		Details:   fmt.Sprintf(details, args),
// 		ErrorType: errType,
// 		// ErrorDocUrl: baseErrDocsUrl + errType, //could be a map somwhere, just an example for now
// 	}

// 	return c.JSON(code, h)
// }

func respErrorFromHttpError(c echo.Context, err *HttpError) error {
	h := HttpError{
		Instance:  c.Request().RequestURI,
		IsError:   true,
		Code:      err.Code,
		Message:   err.Message,
		Details:   err.Details,
		ErrorType: err.ErrorType,
		// ErrorDocUrl: baseErrDocsUrl + err.ErrorType,
	}
	return c.JSON(err.Code, h)
}

type HttpSuccess struct {
	Code    int         `json:"code" example:"200"`
	IsError bool        `json:"isError" example:"false"`
	Message string      `json:"message" example:"OK"`
	Data    interface{} `json:"data,omitempty"`
}

func respSuccess(c echo.Context, code int, message string, data ...interface{}) error {
	h := HttpSuccess{
		Code:    code,
		IsError: false,
		Message: message,
	}

	if len(data) > 0 {
		h.Data = data[0]
	}

	return c.JSON(code, h)
}

// func respSuccessf(c echo.Context, code int, message string, args ...string) error {
// 	h := HttpSuccess{
// 		Code:    code,
// 		IsError: false,
// 		Message: fmt.Sprintf(message, args),
// 	}

// 	return c.JSON(code, h)
// }
