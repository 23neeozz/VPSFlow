# Auth Service

Authentication and session management for VPSFlow.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | Register user |
| POST | `/api/v1/auth/login` | Login with email/password |
| POST | `/api/v1/auth/refresh` | Rotate refresh token |
| POST | `/api/v1/auth/logout` | Revoke session (Bearer required) |
| POST | `/api/v1/auth/mfa/verify` | Verify MFA challenge |
| GET | `/healthz` | Liveness |
| GET | `/readyz` | Readiness |
| GET | `/metrics` | Prometheus metrics |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| AUTH_HTTP_ADDR | `:8081` | HTTP listen address |
| AUTH_DATABASE_URL | `postgres://...` | PostgreSQL connection |
| AUTH_MIGRATIONS_PATH | `migrations` | SQL migrations directory |
| JWT_SIGNING_KEY | (required) | HS256 signing key (min 32 chars) |
| AUTH_ACCESS_TOKEN_TTL | `15m` | Access token lifetime |
| AUTH_REFRESH_TOKEN_TTL | `168h` | Refresh token lifetime |

## Run Locally

```bash
# Start infrastructure (PostgreSQL)
docker compose -f platform/docker/docker-compose.yml up -d postgres

# Copy env and run from service directory
cp .env.example .env
cd services/auth
go run ./cmd/auth
```

## Via Gateway

Gateway proxies `/api/v1/auth/*` to this service when `AUTH_SERVICE_URL` is configured.

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@vpsflow.local","password":"SecurePass123!","name":"Admin"}'
```

## Contract

- [OpenAPI](../../contracts/openapi/auth-v1.yaml)
