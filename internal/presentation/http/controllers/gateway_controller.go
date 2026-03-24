// internal/presentation/http/controllers/gateway_controller.go
package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/farmanexo/api-gateway/internal/presentation/dto/responses"
	"github.com/farmanexo/api-gateway/internal/shared/common"
	"go.uber.org/zap"
)

// GatewayController maneja los endpoints propios del API Gateway
type GatewayController struct {
	logger *zap.Logger
}

// NewGatewayController crea un nuevo controller del gateway
func NewGatewayController(logger *zap.Logger) *GatewayController {
	return &GatewayController{
		logger: logger,
	}
}

// HealthCheck godoc
// @Summary      Health check
// @Description  Verifica el estado del API Gateway
// @Tags         Health
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.ApiResponse[responses.HealthResponse]  "Gateway saludable"
// @Router       /health [get]
func (c *GatewayController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	health := responses.HealthResponse{
		Status:  "healthy",
		Service: "api-gateway",
		Version: "1.0.0",
	}

	resp := common.OkResponse(health)
	c.respondJSON(w, resp)
}

// NotFound handler para rutas no encontradas
func (c *GatewayController) NotFound(w http.ResponseWriter, r *http.Request) {
	resp := common.NotFoundResponse[responses.EmptyResponse](
		"Ruta no encontrada en el API Gateway: " + r.URL.Path,
	)
	c.respondJSON(w, resp)
}

// MethodNotAllowed handler para métodos no permitidos
func (c *GatewayController) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	resp := common.BadRequestResponse[responses.EmptyResponse](
		"VAL_001",
		"Método " + r.Method + " no permitido para " + r.URL.Path,
	)
	c.respondJSON(w, resp)
}

// respondJSON envía una respuesta JSON con el status code correcto
func (c *GatewayController) respondJSON(w http.ResponseWriter, response interface{}) {
	statusCode := http.StatusOK

	if resp, ok := response.(interface{ GetHttpStatus() *int }); ok {
		if httpStatus := resp.GetHttpStatus(); httpStatus != nil {
			statusCode = *httpStatus
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		c.logger.Error("Error codificando respuesta JSON", zap.Error(err))
	}
}
