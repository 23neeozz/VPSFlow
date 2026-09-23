# Tenant Service

Organization, project, and membership management.

**Status**: Planned — Sprint 1

See [Identity Architecture](../../docs/architecture/identity.md) for full specification.

## Planned Endpoints

- `GET/POST /api/v1/organizations`
- `GET /api/v1/organizations/{org_id}`
- `GET/POST /api/v1/organizations/{org_id}/members`
- `GET/POST /api/v1/organizations/{org_id}/projects`

## Contract

- [OpenAPI](../../contracts/openapi/tenant-v1.yaml)
