# IAM Service

Identity and Access Management — roles, permissions, and policy evaluation.

**Status**: Planned — Sprint 1

See [Identity Architecture](../../docs/architecture/identity.md) for full specification.

## Planned Endpoints

- `GET /api/v1/iam/roles`
- `POST /api/v1/iam/roles`
- `GET /api/v1/iam/permissions`
- `PUT /api/v1/iam/users/{user_id}/roles`
- `POST /api/v1/iam/policy/evaluate`

## Contract

- [OpenAPI](../../contracts/openapi/iam-v1.yaml)
