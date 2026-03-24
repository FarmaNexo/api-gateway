// internal/shared/constants/message_codes.go
package constants

// MessageCode contiene todos los códigos de respuesta del sistema
type MessageCode string

const (
	// Success codes
	CodeSuccess        MessageCode = "SUCCESS_001"
	CodeCreatedSuccess MessageCode = "SUCCESS_002"
	CodeUpdatedSuccess MessageCode = "SUCCESS_003"
	CodeDeletedSuccess MessageCode = "SUCCESS_004"

	// Gateway codes
	CodeGatewayHealthy    MessageCode = "GW_001"
	CodeGatewayProxyOK    MessageCode = "GW_002"
	CodeGatewayRouted     MessageCode = "GW_003"

	// Validation errors
	CodeValidationError  MessageCode = "VAL_001"
	CodeRequiredField    MessageCode = "VAL_005"

	// Authentication errors
	CodeUnauthorized       MessageCode = "AUTH_ERR_001"
	CodeInvalidToken       MessageCode = "AUTH_ERR_002"
	CodeTokenExpired       MessageCode = "AUTH_ERR_003"
	CodeMissingToken       MessageCode = "AUTH_ERR_007"

	// Gateway errors
	CodeGatewayError       MessageCode = "GW_ERR_001"
	CodeServiceUnavailable MessageCode = "GW_ERR_002"
	CodeCircuitOpen        MessageCode = "GW_ERR_003"
	CodeUpstreamTimeout    MessageCode = "GW_ERR_004"
	CodeRouteNotFound      MessageCode = "GW_ERR_005"
	CodeBadGateway         MessageCode = "GW_ERR_006"

	// Rate limiting
	CodeRateLimitExceeded MessageCode = "RATE_001"

	// Business errors
	CodeResourceNotFound MessageCode = "BUS_003"

	// System errors
	CodeInternalError MessageCode = "SYS_001"
)

// MessageDescription contiene las descripciones predefinidas
var MessageDescription = map[MessageCode]string{
	// Success
	CodeSuccess:        "Operación exitosa",
	CodeCreatedSuccess: "Recurso creado exitosamente",
	CodeUpdatedSuccess: "Recurso actualizado exitosamente",
	CodeDeletedSuccess: "Recurso eliminado exitosamente",

	// Gateway
	CodeGatewayHealthy: "API Gateway operativo",
	CodeGatewayProxyOK: "Request proxied exitosamente",
	CodeGatewayRouted:  "Request ruteado al servicio",

	// Validation
	CodeValidationError: "Error de validación",
	CodeRequiredField:   "Campo requerido",

	// Auth errors
	CodeUnauthorized: "No autorizado",
	CodeInvalidToken: "Token inválido",
	CodeTokenExpired: "Token expirado",
	CodeMissingToken: "Token de autenticación requerido",

	// Gateway errors
	CodeGatewayError:       "Error en el API Gateway",
	CodeServiceUnavailable: "Servicio no disponible temporalmente",
	CodeCircuitOpen:        "Servicio temporalmente deshabilitado por fallos consecutivos",
	CodeUpstreamTimeout:    "Tiempo de espera agotado con el servicio downstream",
	CodeRouteNotFound:      "Ruta no encontrada en el API Gateway",
	CodeBadGateway:         "Error de comunicación con el servicio downstream",

	// Rate limiting
	CodeRateLimitExceeded: "Demasiadas solicitudes. Intente nuevamente más tarde",

	// Business
	CodeResourceNotFound: "Recurso no encontrado",

	// System
	CodeInternalError: "Error interno del servidor",
}

// GetDescription retorna la descripción del código
func GetDescription(code MessageCode) string {
	if desc, ok := MessageDescription[code]; ok {
		return desc
	}
	return "Descripción no disponible"
}
