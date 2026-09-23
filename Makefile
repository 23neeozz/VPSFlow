# VPSFlow Makefile

.PHONY: help infra-up infra-down test build-gateway build-auth run-gateway run-auth tidy

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

infra-up: ## Start local infrastructure (PostgreSQL, Redis, NATS, MinIO, observability)
	docker compose -f platform/docker/docker-compose.yml up -d

infra-down: ## Stop local infrastructure
	docker compose -f platform/docker/docker-compose.yml down

infra-logs: ## Follow infrastructure logs
	docker compose -f platform/docker/docker-compose.yml logs -f

test: ## Run all Go tests
	cd libs/go/config && go test -race -count=1 ./...
	cd libs/go/errors && go test -race -count=1 ./...
	cd libs/go/ids && go test -race -count=1 ./...
	cd libs/go/observability && go test -race -count=1 ./...
	cd libs/go/httpx && go test -race -count=1 ./...
	cd libs/go/security && go test -race -count=1 ./...
	cd services/gateway && go test -race -count=1 ./...
	cd services/auth && go test -race -count=1 ./...

build-gateway: ## Build gateway binary
	cd services/gateway && go build -o ../../bin/gateway ./cmd/gateway

build-auth: ## Build auth binary
	cd services/auth && go build -o ../../bin/auth ./cmd/auth

run-gateway: ## Run gateway service locally
	cd services/gateway && go run ./cmd/gateway

run-auth: ## Run auth service locally
	cd services/auth && go run ./cmd/auth

tidy: ## Tidy all Go modules
	cd libs/go/config && go mod tidy
	cd libs/go/errors && go mod tidy
	cd libs/go/ids && go mod tidy
	cd libs/go/observability && go mod tidy
	cd libs/go/httpx && go mod tidy
	cd libs/go/security && go mod tidy
	cd services/gateway && go mod tidy
	cd services/auth && go mod tidy

vet: ## Run go vet on all modules
	go vet ./libs/go/config/...
	go vet ./libs/go/errors/...
	go vet ./libs/go/ids/...
	go vet ./libs/go/observability/...
	go vet ./libs/go/httpx/...
	go vet ./libs/go/security/...
	go vet ./services/gateway/...
	go vet ./services/auth/...

docker-build-gateway: ## Build gateway Docker image
	docker build -f services/gateway/Dockerfile -t vpsflow/gateway:latest .

docker-build-auth: ## Build auth Docker image
	docker build -f services/auth/Dockerfile -t vpsflow/auth:latest .
