# Makefile para API Gateway

.PHONY: help build run test clean docker-build docker-run lint swagger

# Variables
SERVICE_NAME=api-gateway
BINARY_NAME=api-gateway
DOCKER_IMAGE=farmanexo/$(SERVICE_NAME):latest
ENV?=local

help: ## Muestra esta ayuda
	@echo "Comandos disponibles:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install: ## Instala las dependencias
	@echo "Instalando dependencias..."
	go mod download
	go mod tidy
	go install github.com/swaggo/swag/cmd/swag@latest

build: ## Compila el binario
	@echo "Compilando $(BINARY_NAME)..."
	go build -o bin/$(BINARY_NAME) cmd/server/main.go

run: swagger ## Ejecuta el servicio
	@echo "Ejecutando $(SERVICE_NAME) en modo $(ENV)..."
	ENV=$(ENV) go run cmd/server/main.go

dev: swagger ## Ejecuta en modo local con auto-reload
	@echo "Ejecutando en modo local..."
	ENV=local go run cmd/server/main.go

swagger: ## Genera documentación Swagger
	@echo "Generando documentación Swagger..."
	swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
	@echo "Swagger generado en /docs"

test: ## Ejecuta los tests
	@echo "Ejecutando tests..."
	go test -v -race -cover ./...

test-coverage: ## Ejecuta tests con coverage
	@echo "Ejecutando tests con coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generado: coverage.html"

lint: ## Ejecuta el linter
	@echo "Ejecutando linter..."
	golangci-lint run ./...

clean: ## Limpia archivos generados
	@echo "Limpiando..."
	rm -rf bin/
	rm -rf docs/
	rm -f coverage.out coverage.html

# Docker commands
docker-build: ## Construye la imagen Docker
	@echo "Construyendo imagen Docker..."
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Ejecuta el contenedor Docker
	@echo "Ejecutando contenedor..."
	docker run -p 8080:8080 --env-file .env.$(ENV) $(DOCKER_IMAGE)

docker-push: ## Sube la imagen a Docker Hub
	@echo "Subiendo imagen..."
	docker push $(DOCKER_IMAGE)

# Development helpers
watch: ## Watch mode con air
	@echo "Iniciando watch mode..."
	air

format: ## Formatea el código
	@echo "Formateando código..."
	go fmt ./...
	goimports -w .

# Deployment
deploy-dev: ## Deploy a desarrollo
	@echo "Deploying to development..."

deploy-prod: ## Deploy a producción
	@echo "Deploying to production..."

.DEFAULT_GOAL := help
