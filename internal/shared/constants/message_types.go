// internal/shared/constants/message_types.go
package constants

// MessageType define los tipos de mensaje posibles
type MessageType string

const (
	MessageTypeInformation MessageType = "INFORMATION"
	MessageTypeWarning     MessageType = "WARNING"
	MessageTypeError       MessageType = "ERROR"
	MessageTypeSuccess     MessageType = "SUCCESS"
)

// IsValid verifica si el tipo de mensaje es válido
func (mt MessageType) IsValid() bool {
	switch mt {
	case MessageTypeInformation, MessageTypeWarning, MessageTypeError, MessageTypeSuccess:
		return true
	}
	return false
}

// String retorna la representación en string del MessageType
func (mt MessageType) String() string {
	return string(mt)
}
