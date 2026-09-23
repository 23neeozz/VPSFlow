# ADR-003: API Versioning Strategy

## Status

Accepted

## Context

BossCloud exposes public REST APIs to customers and SDK/CLI integrations. We need a versioning strategy that allows evolution without breaking consumers.

## Decision

- URI versioning: `/api/v1/`, `/api/v2/`
- Breaking changes require new major version
- Non-breaking additions (new fields, new endpoints) within same version
- Deprecation window: minimum 2 major versions overlap
- Sunset headers on deprecated endpoints: `Sunset`, `Deprecation`
- OpenAPI spec per version in `contracts/openapi/`

## Consequences

### Positive
- Clear consumer expectations
- Simple routing at gateway level
- Easy documentation per version

### Negative
- Multiple API versions to maintain during transition
- Gateway routing complexity increases

## Alternatives Considered

| Alternative | Pros | Cons | Why Rejected |
|-------------|------|------|--------------|
| Header versioning | Clean URIs | Harder to discover, test, document | Poor DX for public API |
| Query param versioning | Flexible | Non-standard, cache issues | Not industry practice |

## References

- [Gateway OpenAPI](../../contracts/openapi/gateway-v1.yaml)
