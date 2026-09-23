# Getting Started

## Prerequisites

- Go 1.23+
- Docker Desktop (or Docker Engine + Compose)
- Git

## 1. Clone and Configure

```bash
git clone <repository-url> bosscloud
cd bosscloud
cp .env.example .env
```

## 2. Start Infrastructure

```bash
docker compose -f platform/docker/docker-compose.yml up -d
```

This starts:
- PostgreSQL (port 5432) with per-service databases
- Redis (port 6379)
- NATS JetStream (port 4222)
- MinIO (ports 9000/9001)
- Prometheus (port 9090)
- Grafana (port 3001, admin/bosscloud_dev)
- OpenTelemetry Collector (ports 4317/4318)

## 3. Run Services

```bash
# Terminal 1 - Auth (requires PostgreSQL)
cd services/auth
go run ./cmd/auth

# Terminal 2 - Gateway
cd services/gateway
go run ./cmd/gateway
```

## 4. Verify

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/api/v1

# Register via gateway proxy
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@bosscloud.local","password":"SecurePass123!","name":"Admin"}'
```

## 5. Run Tests

```bash
# All Go tests
cd libs/go/config && go test ./...
cd ../errors && go test ./...
cd ../observability && go test ./...
cd ../httpx && go test ./...
cd ../../../services/gateway && go test ./...
```

## Development Workflow

1. Create feature branch from `develop`
2. Implement with tests (no TODOs, no stubs)
3. Run local tests and lint
4. Open PR to `develop`
5. CI must pass (lint, test, contract validation, security scan)
6. Merge after review

## Useful URLs (Local)

| Service    | URL                          |
|------------|------------------------------|
| Gateway    | http://localhost:8080        |
| Prometheus | http://localhost:9090        |
| Grafana    | http://localhost:3001        |
| MinIO UI   | http://localhost:9001        |
| NATS       | http://localhost:8222        |
