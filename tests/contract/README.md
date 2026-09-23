# Contract Tests

Cross-service API and event contract validation.

## Structure

- `openapi/` — OpenAPI contract tests against running services
- `grpc/` — gRPC contract tests with protobuf validation
- `events/` — AsyncAPI event schema validation

## Running

Contract validation runs automatically in CI via `.github/workflows/ci.yml`.

Manual validation:

```bash
# OpenAPI
npx @redocly/cli lint contracts/openapi/*.yaml
```
