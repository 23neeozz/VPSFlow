# ADR-001: Monorepo Structure

## Status

Accepted

## Context

BossCloud requires coordinated development across 25+ microservices, shared libraries, frontend applications, infrastructure code, and API contracts. We need a repository structure that supports independent service deployment while enabling code sharing and consistent standards.

## Decision

Adopt a **monorepo** with the following top-level structure:

- `services/` — Go microservices (independently deployable)
- `agents/` — Hypervisor agent
- `frontend/` — Next.js applications
- `libs/` — Shared Go and TypeScript libraries
- `contracts/` — OpenAPI and AsyncAPI specifications
- `proto/` — gRPC protobuf definitions
- `platform/` — Docker, Helm, Terraform, observability configs
- `docs/` — Architecture, ADRs, standards
- `tests/` — Cross-service contract, E2E, performance tests

Go workspace (`go.work`) manages multi-module Go dependencies.

## Consequences

### Positive
- Single source of truth for contracts and shared libraries
- Atomic changes across services and contracts
- Unified CI/CD pipeline
- Consistent standards enforcement

### Negative
- Repository size grows over time
- Requires disciplined module boundaries
- CI must be optimized (path-based triggers)

### Neutral
- Each service still has its own Dockerfile and deployment lifecycle

## Alternatives Considered

| Alternative | Pros | Cons | Why Rejected |
|-------------|------|------|--------------|
| Polyrepo (repo per service) | Independent scaling of repos | Contract drift, shared lib versioning pain | Too many services for early stage |
| Monorepo with Bazel | Advanced build caching | Complexity overhead for Go/TS stack | Over-engineering for current team size |

## References

- [Architecture Overview](../architecture/overview.md)
