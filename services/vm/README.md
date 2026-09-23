# VM Service

Virtual machine lifecycle management.

**Status**: Planned — Sprint 3

See [Compute Architecture](../../docs/architecture/compute.md) for full specification.

## Planned Operations

- Create, start, stop, delete, reinstall VM
- Idempotent operations via `Idempotency-Key`
- Cloud-init configuration
- State machine with event publication
