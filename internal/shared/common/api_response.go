// internal/shared/common/api_response.go
package common

import (
	"time"

	"github.com/farmanexo/api-gateway/internal/shared/constants"
	"github.com/google/uuid"
)

// ApiResponse es la respuesta estándar de la API (equivalente a C#)
// Usa generics [T any] para ser type-safe
type ApiResponse[T any] struct {
	httpStatus *int  `json:"-"` // No se serializa en JSON
	Meta       *Meta `json:"meta"`
	Data       *T    `json:"datos"`
}

// Meta contiene metadata de la respuesta
type Meta struct {
	Messages      []ResponseMessage `json:"mensajes"`
	IdTransaction string            `json:"idTransaccion"`
	Result        bool              `json:"resultado"`
	Timestamp     string            `json:"timestamp"`
}

// ResponseMessage representa un mensaje individual
type ResponseMessage struct {
	Code    string `json:"codigo"`
	Message string `json:"mensaje"`
	Type    string `json:"tipo"`
}

// NewApiResponse crea una nueva instancia de ApiResponse
func NewApiResponse[T any]() *ApiResponse[T] {
	return &ApiResponse[T]{
		Meta: &Meta{
			Messages:      make([]ResponseMessage, 0),
			IdTransaction: uuid.New().String(),
			Result:        true,
			Timestamp:     time.Now().Format("20060102 150405"),
		},
		Data: nil,
	}
}

// ========================================
// HTTP STATUS METHODS
// ========================================

// SetHttpStatus establece el código HTTP de la respuesta
func (r *ApiResponse[T]) SetHttpStatus(statusCode int) {
	r.httpStatus = &statusCode
}

// GetHttpStatus obtiene el código HTTP (nil si no está establecido)
func (r *ApiResponse[T]) GetHttpStatus() *int {
	return r.httpStatus
}

// GetHttpStatusOrDefault obtiene el HTTP status o retorna el default
func (r *ApiResponse[T]) GetHttpStatusOrDefault(defaultStatus int) int {
	if r.httpStatus != nil {
		return *r.httpStatus
	}
	return defaultStatus
}

// ========================================
// MESSAGE METHODS
// ========================================

// AddMessage agrega un mensaje con tipo INFORMATION
func (r *ApiResponse[T]) AddMessage(code constants.MessageCode, message string) {
	r.Meta.AddMessage(
		string(code),
		message,
		string(constants.MessageTypeInformation),
	)
}

// AddMessageWithType agrega un mensaje con tipo específico
func (r *ApiResponse[T]) AddMessageWithType(
	code constants.MessageCode,
	message string,
	messageType constants.MessageType,
) {
	r.Meta.AddMessage(string(code), message, string(messageType))
}

// AddError agrega un mensaje de error
func (r *ApiResponse[T]) AddError(code constants.MessageCode, message string) {
	r.Meta.AddError(string(code), message)
}

// AddErrorSimple agrega un error con código genérico
func (r *ApiResponse[T]) AddErrorSimple(message string) {
	r.Meta.AddErrorSimple(message)
}

// AddSuccessMessage agrega mensaje de éxito genérico
func (r *ApiResponse[T]) AddSuccessMessage() {
	r.Meta.AddMessage(
		string(constants.CodeSuccess),
		constants.GetDescription(constants.CodeSuccess),
		string(constants.MessageTypeInformation),
	)
}

// ========================================
// DATA METHODS
// ========================================

// SetData establece los datos de la respuesta
func (r *ApiResponse[T]) SetData(data T) {
	r.Data = &data
}

// GetData obtiene los datos de la respuesta
func (r *ApiResponse[T]) GetData() T {
	if r.Data != nil {
		return *r.Data
	}
	var zero T
	return zero
}

// ========================================
// VALIDATION METHODS
// ========================================

// IsValid verifica si la respuesta es válida (sin errores)
func (r *ApiResponse[T]) IsValid() bool {
	return r.Meta != nil && r.Meta.Result
}

// HasErrors verifica si hay errores
func (r *ApiResponse[T]) HasErrors() bool {
	return !r.IsValid()
}

// ========================================
// META METHODS
// ========================================

// AddMessage agrega un mensaje a Meta
func (m *Meta) AddMessage(code, message, messageType string) {
	m.Messages = append(m.Messages, ResponseMessage{
		Code:    code,
		Message: message,
		Type:    messageType,
	})

	if messageType == string(constants.MessageTypeError) {
		m.Result = false
	}
}

// AddError agrega un error a Meta
func (m *Meta) AddError(code, message string) {
	m.Result = false
	m.Messages = append(m.Messages, ResponseMessage{
		Code:    code,
		Message: message,
		Type:    string(constants.MessageTypeError),
	})
}

// AddErrorSimple agrega un error simple con código genérico
func (m *Meta) AddErrorSimple(message string) {
	m.AddError(
		string(constants.CodeInternalError),
		message,
	)
}
