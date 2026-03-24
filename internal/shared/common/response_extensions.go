// internal/shared/common/response_extensions.go
package common

import (
	"github.com/farmanexo/api-gateway/internal/shared/constants"
)

// ========================================
// FACTORY METHODS (Shortcuts)
// ========================================

// OkResponse crea una respuesta 200 OK con datos
func OkResponse[T any](data T) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetData(data)
	resp.AddSuccessMessage()
	resp.SetHttpStatus(constants.StatusOK.Int())
	return resp
}

// BadRequestResponse crea una respuesta 400 Bad Request
func BadRequestResponse[T any](code constants.MessageCode, message string) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusBadRequest.Int())
	resp.AddError(code, message)
	return resp
}

// UnauthorizedResponse crea una respuesta 401 Unauthorized
func UnauthorizedResponse[T any](message string) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusUnauthorized.Int())
	resp.AddError(constants.CodeUnauthorized, message)
	return resp
}

// ForbiddenResponse crea una respuesta 403 Forbidden
func ForbiddenResponse[T any](message string) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusForbidden.Int())
	resp.AddError(constants.CodeUnauthorized, message)
	return resp
}

// NotFoundResponse crea una respuesta 404 Not Found
func NotFoundResponse[T any](message string) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusNotFound.Int())
	resp.AddError(constants.CodeRouteNotFound, message)
	return resp
}

// TooManyRequestsResponse crea una respuesta 429 Too Many Requests
func TooManyRequestsResponse[T any](message string) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusTooManyRequests.Int())
	resp.AddError(constants.CodeRateLimitExceeded, message)
	return resp
}

// InternalServerErrorResponse crea una respuesta 500
func InternalServerErrorResponse[T any](message string) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusInternalServerError.Int())
	resp.AddError(constants.CodeInternalError, message)
	return resp
}

// BadGatewayResponse crea una respuesta 502 Bad Gateway
func BadGatewayResponse[T any](message string) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusBadGateway.Int())
	resp.AddError(constants.CodeBadGateway, message)
	return resp
}

// ServiceUnavailableResponse crea una respuesta 503 Service Unavailable
func ServiceUnavailableResponse[T any](message string) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusServiceUnavailable.Int())
	resp.AddError(constants.CodeServiceUnavailable, message)
	return resp
}

// GatewayTimeoutResponse crea una respuesta 504 Gateway Timeout
func GatewayTimeoutResponse[T any](message string) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusGatewayTimeout.Int())
	resp.AddError(constants.CodeUpstreamTimeout, message)
	return resp
}
