// internal/infrastructure/proxy/reverse_proxy.go
package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ServiceProxy encapsula un reverse proxy hacia un microservicio
type ServiceProxy struct {
	Name       string
	TargetURL  *url.URL
	Proxy      *httputil.ReverseProxy
	Timeout    time.Duration
	logger     *zap.Logger
}

// headersToStrip headers internos que no se deben reenviar al cliente
var headersToStrip = []string{
	"X-Powered-By",
	"Server",
}

// NewServiceProxy crea un nuevo proxy hacia un servicio downstream
func NewServiceProxy(name, targetURL string, timeout time.Duration, logger *zap.Logger) (*ServiceProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL for service %s: %w", name, err)
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = transport

	sp := &ServiceProxy{
		Name:      name,
		TargetURL: target,
		Proxy:     proxy,
		Timeout:   timeout,
		logger:    logger,
	}

	// Director: modifica el request antes de enviarlo al downstream
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host

		// Propagar X-Correlation-ID
		if corrID := req.Header.Get("X-Correlation-ID"); corrID != "" {
			req.Header.Set("X-Correlation-ID", corrID)
		}

		// Agregar header indicando que viene del gateway
		req.Header.Set("X-Forwarded-By", "farmanexo-api-gateway")
	}

	// ModifyResponse: modifica la respuesta antes de enviarla al cliente
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Eliminar headers internos del downstream
		for _, h := range headersToStrip {
			resp.Header.Del(h)
		}
		return nil
	}

	// ErrorHandler: maneja errores de proxy
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Error("Proxy error",
			zap.String("service", name),
			zap.String("target", target.String()),
			zap.String("path", r.URL.Path),
			zap.Error(err),
		)

		statusCode := http.StatusBadGateway
		message := "Error de comunicación con el servicio"

		if isTimeout(err) {
			statusCode = http.StatusGatewayTimeout
			message = "Tiempo de espera agotado con el servicio downstream"
		}

		if isConnectionRefused(err) {
			statusCode = http.StatusServiceUnavailable
			message = "Servicio no disponible temporalmente"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		fmt.Fprintf(w, `{"meta":{"mensajes":[{"codigo":"GW_ERR_006","mensaje":"%s","tipo":"ERROR"}],"idTransaccion":"","resultado":false,"timestamp":""},"datos":null}`, message)
	}

	return sp, nil
}

// ServeHTTP implementa http.Handler para usar con circuit breaker
func (sp *ServiceProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if sp.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, sp.Timeout)
		defer cancel()
	}
	sp.Proxy.ServeHTTP(w, r.WithContext(ctx))
}

// isTimeout verifica si el error es un timeout
func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	netErr, ok := err.(net.Error)
	return ok && netErr.Timeout()
}

// isConnectionRefused verifica si el error es connection refused
func isConnectionRefused(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "no such host")
}
