package uaa

import (
	"encoding/json"
	"net/http"

	"code.cloudfoundry.org/cli/api/shared"
)

// errorWrapper is the wrapper that converts responses with 4xx and 5xx status
// codes to an error.
type errorWrapper struct {
	connection shared.Connection
}

// NewErrorWrapper returns a new error wrapper.
func NewErrorWrapper() *errorWrapper {
	return new(errorWrapper)
}

// Make converts RawHTTPStatusError, which represents responses with 4xx and
// 5xx status codes, to specific errors.
func (e *errorWrapper) Make(request *shared.Request, passedResponse shared.Response) error {
	err := e.connection.Make(request, passedResponse)

	if rawHTTPStatusErr, ok := err.(RawHTTPStatusError); ok {
		return convert(rawHTTPStatusErr)
	}

	return err
}

// Wrap wraps a UAA connection in this error handling wrapper.
func (e *errorWrapper) Wrap(innerconnection shared.Connection) shared.Connection {
	e.connection = innerconnection
	return e
}

func convert(rawHTTPStatusErr RawHTTPStatusError) error {
	// Try to unmarshal the raw http status error into a UAA error. If
	// unmarshaling fails, return the raw error.
	var uaaErrorResponse UAAErrorResponse
	err := json.Unmarshal(rawHTTPStatusErr.RawResponse, &uaaErrorResponse)
	if err != nil {
		return rawHTTPStatusErr
	}

	switch rawHTTPStatusErr.StatusCode {
	case http.StatusBadRequest: // 400
		if uaaErrorResponse.Type == "invalid_scim_resource" {
			return InvalidSCIMResourceError{Message: uaaErrorResponse.Description}
		}
		return rawHTTPStatusErr
	case http.StatusUnauthorized: // 401
		if uaaErrorResponse.Type == "invalid_token" {
			return InvalidAuthTokenError{Message: uaaErrorResponse.Description}
		}
		if uaaErrorResponse.Type == "unauthorized" {
			return UnauthorizedError{Message: uaaErrorResponse.Description}
		}
		return rawHTTPStatusErr
	case http.StatusForbidden: // 403
		if uaaErrorResponse.Type == "insufficient_scope" {
			return InsufficientScopeError{Message: uaaErrorResponse.Description}
		}
		return rawHTTPStatusErr
	case http.StatusConflict: // 409
		return ConflictError{Message: uaaErrorResponse.Description}
	default:
		return rawHTTPStatusErr
	}
}
