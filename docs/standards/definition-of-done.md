# Definition of Done

A sprint item is considered **Done** only when ALL criteria below are met.

## Code Quality

- [ ] Implementation complete — no stubs, no TODOs, no empty functions
- [ ] Follows Clean Architecture (domain isolated from infrastructure)
- [ ] Follows naming conventions and coding standards
- [ ] Dependency injection via constructors (no global state)
- [ ] Error handling uses canonical `AppError` type

## Testing

- [ ] Unit tests for domain and usecase layers (>85% coverage)
- [ ] Integration tests for adapters (DB, messaging, HTTP)
- [ ] Contract tests if API/events changed
- [ ] All tests pass locally and in CI

## Security

- [ ] Input validation on all endpoints
- [ ] Authentication/authorization enforced where required
- [ ] No secrets in code or logs
- [ ] Security scan passes (govulncheck, dependency audit)

## Observability

- [ ] Structured logging with consistent fields
- [ ] Prometheus metrics exposed on `/metrics`
- [ ] OpenTelemetry tracing instrumented
- [ ] Health (`/healthz`) and readiness (`/readyz`) endpoints

## Documentation

- [ ] Service README updated
- [ ] API contract updated (OpenAPI/gRPC proto) if applicable
- [ ] ADR written for architectural decisions
- [ ] Environment variables documented

## Operations

- [ ] Dockerfile builds successfully
- [ ] Database migrations included and tested (forward + backward)
- [ ] Configuration via environment variables only
- [ ] Graceful shutdown implemented

## Review

- [ ] Code review completed and approved
- [ ] CI pipeline green
- [ ] Performance acceptable (no regression on p95)
- [ ] Risks documented and mitigations identified
