# Gateway Service

API Gateway for VPSFlow. Entry point for all public REST and WebSocket traffic.

## Responsibilities

- Request routing to downstream services
- Authentication token validation (delegated to auth service)
- Rate limiting and request validation
- API versioning (`/api/v1`)
- Health, readiness, and metrics endpoints

## Endpoints

| Method | Path       | Description              |
|--------|------------|--------------------------|
| GET    | /healthz   | Liveness probe           |
| GET    | /readyz    | Readiness probe          |
| GET    | /metrics   | Prometheus metrics       |
| GET    | /api/v1    | API root / service info  |

## Configuration

| Variable                  | Default     | Description                    |
|---------------------------|-------------|--------------------------------|
| GATEWAY_HTTP_ADDR         | :8080       | HTTP listen address            |
| GATEWAY_ENV               | development | Environment name               |
| GATEWAY_LOG_LEVEL         | info        | Log level (debug/info/warn)    |
| GATEWAY_SERVICE_NAME      | gateway     | Service identifier             |
| GATEWAY_SERVICE_VERSION   | 0.1.0       | Service version                |
| OTEL_EXPORTER_OTLP_ENDPOINT | (empty)   | OTLP collector endpoint        |

## Run Locally

```bash
go run ./cmd/gateway
```

## Run Tests

```bash
go test ./...
```

## Docker

```bash
docker build -f Dockerfile -t vpsflow/gateway:latest ../..
```
