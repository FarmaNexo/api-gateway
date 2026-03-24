// internal/shared/constants/http_status.go
package constants

// HTTPStatusCode define los códigos HTTP estándar
type HTTPStatusCode int

const (
	// Success 2xx
	StatusOK        HTTPStatusCode = 200
	StatusCreated   HTTPStatusCode = 201
	StatusAccepted  HTTPStatusCode = 202
	StatusNoContent HTTPStatusCode = 204

	// Client Errors 4xx
	StatusBadRequest          HTTPStatusCode = 400
	StatusUnauthorized        HTTPStatusCode = 401
	StatusForbidden           HTTPStatusCode = 403
	StatusNotFound            HTTPStatusCode = 404
	StatusMethodNotAllowed    HTTPStatusCode = 405
	StatusConflict            HTTPStatusCode = 409
	StatusUnprocessableEntity HTTPStatusCode = 422
	StatusTooManyRequests     HTTPStatusCode = 429

	// Server Errors 5xx
	StatusInternalServerError HTTPStatusCode = 500
	StatusNotImplemented      HTTPStatusCode = 501
	StatusBadGateway          HTTPStatusCode = 502
	StatusServiceUnavailable  HTTPStatusCode = 503
	StatusGatewayTimeout      HTTPStatusCode = 504
)

// Int retorna el código como int
func (h HTTPStatusCode) Int() int {
	return int(h)
}

// String retorna la descripción del código
func (h HTTPStatusCode) String() string {
	descriptions := map[HTTPStatusCode]string{
		StatusOK:                  "OK",
		StatusCreated:             "Created",
		StatusAccepted:            "Accepted",
		StatusNoContent:           "No Content",
		StatusBadRequest:          "Bad Request",
		StatusUnauthorized:        "Unauthorized",
		StatusForbidden:           "Forbidden",
		StatusNotFound:            "Not Found",
		StatusMethodNotAllowed:    "Method Not Allowed",
		StatusConflict:            "Conflict",
		StatusUnprocessableEntity: "Unprocessable Entity",
		StatusTooManyRequests:     "Too Many Requests",
		StatusInternalServerError: "Internal Server Error",
		StatusNotImplemented:      "Not Implemented",
		StatusBadGateway:          "Bad Gateway",
		StatusServiceUnavailable:  "Service Unavailable",
		StatusGatewayTimeout:      "Gateway Timeout",
	}

	if desc, ok := descriptions[h]; ok {
		return desc
	}
	return "Unknown Status"
}
