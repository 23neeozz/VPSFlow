# Coding Standards

## General Rules

1. **No temporary code** — every function must be complete and production-ready.
2. **No TODO comments** — implement fully or create a tracked issue.
3. **No empty functions** — all exported functions must have real behavior.
4. **Test everything** — unit tests for domain/usecase, integration for adapters.

## Go Standards

### Package Structure (per service)
```
services/<name>/
  cmd/<name>/          # Entry point
  internal/
    domain/            # Entities, value objects, domain rules
    usecase/           # Application logic
    port/              # Interface definitions
    adapter/           # HTTP, gRPC, DB, messaging implementations
    config/            # Service configuration
  migrations/          # SQL migrations
  test/                # Integration tests
```

### Conventions
- Use `context.Context` as first parameter in all I/O functions
- Return errors, never panic in library code
- Use `AppError` from `libs/go/errors` for domain errors
- Structured logging via `slog` with consistent field names
- Constructor-based dependency injection (no global state)

### Error Handling
```go
// Correct
return nil, errors.Wrap(errors.CodeNotFound, "vm not found", err)

// Incorrect
return nil, fmt.Errorf("not found")
```

## TypeScript Standards (Frontend)

- Strict TypeScript (`strict: true`)
- Feature-based module organization
- Server Components where beneficial
- Zod schemas for all API input validation
- TanStack Query for server state, Zustand for minimal client state

## API Standards

- REST: snake_case JSON fields
- Versioning: URI prefix `/api/v1`
- Pagination: cursor-based
- Errors: unified envelope (`code`, `message`, `details`, `trace_id`)
- Idempotency: `Idempotency-Key` header on critical POST operations

## Git Commit Messages

```
<type>(<scope>): <description>

Types: feat, fix, docs, refactor, test, chore, ci
Scope: service name or area (gateway, auth, vm, etc.)
```

## Code Review Checklist

- [ ] Domain logic isolated from infrastructure
- [ ] Tests cover happy path and error cases
- [ ] No secrets in code
- [ ] Observability instrumented (logs, metrics, traces)
- [ ] API contract updated if endpoints changed
- [ ] Migration included if schema changed
- [ ] README updated if behavior changed
