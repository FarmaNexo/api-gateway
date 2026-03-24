// internal/presentation/dto/responses/gateway_response.go
package responses

// HealthResponse respuesta del health check del gateway
type HealthResponse struct {
	Status   string                    `json:"status" example:"healthy"`
	Service  string                    `json:"service" example:"api-gateway"`
	Version  string                    `json:"version" example:"1.0.0"`
	Services map[string]ServiceStatus  `json:"services,omitempty"`
}

// ServiceStatus estado de un servicio downstream
type ServiceStatus struct {
	URL    string `json:"url"`
	Status string `json:"status"`
}

// EmptyResponse representa una respuesta sin datos
type EmptyResponse struct{}
