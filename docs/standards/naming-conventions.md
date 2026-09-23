# Naming Conventions

## Repository & Services

| Element | Convention | Example |
|---------|-----------|---------|
| Service directory | kebab-case | `agent-control` |
| Go package | lowercase, no dashes | `agentcontrol` |
| Docker image | bosscloud/<service> | `bosscloud/gateway` |
| Helm chart | bosscloud-<service> | `bosscloud-gateway` |

## API

| Element | Convention | Example |
|---------|-----------|---------|
| REST resources | plural, kebab-case | `/api/v1/virtual-machines` |
| JSON fields | snake_case | `tenant_id`, `created_at` |
| Query params | snake_case | `?page_size=20` |
| Headers | Pascal-Case with prefix | `X-Request-ID`, `Idempotency-Key` |

## Events

Format: `bosscloud.<bounded_context>.<aggregate>.<event>.v1`

Examples:
- `bosscloud.auth.user.authenticated.v1`
- `bosscloud.vm.instance.created.v1`
- `bosscloud.storage.volume.attached.v1`

## Database

| Element | Convention | Example |
|---------|-----------|---------|
| Tables | snake_case, plural | `virtual_machines` |
| Columns | snake_case | `tenant_id`, `hypervisor_id` |
| Primary keys | `<entity>_id` | `vm_id` |
| Indexes | `idx_<table>_<columns>` | `idx_vms_tenant_id` |
| Foreign keys | `fk_<table>_<ref_table>` | `fk_vms_tenant` |

## Go Code

| Element | Convention | Example |
|---------|-----------|---------|
| Exported types | PascalCase | `VirtualMachine` |
| Unexported | camelCase | `validateInput` |
| Interfaces | noun or -er suffix | `Repository`, `Publisher` |
| Constants | PascalCase or ALL_CAPS for env | `CodeNotFound` |
| Test files | `_test.go` suffix | `vm_test.go` |
| Test functions | `Test<Function>_<Scenario>` | `TestCreateVM_InvalidInput` |

## Identifiers

- Use **ULID** or **UUIDv7** for all primary identifiers
- Never use auto-increment integers for public IDs
- `tenant_id` required on all tenant-scoped resources

## Environment Variables

Format: `<SERVICE>_<SETTING>` in UPPER_SNAKE_CASE

Examples:
- `GATEWAY_HTTP_ADDR`
- `AUTH_JWT_SIGNING_KEY`
- `VM_LIBVIRT_TIMEOUT`

## Proto / gRPC

- Package: `bosscloud.<domain>.v1`
- Services: PascalCase (`VirtualMachineService`)
- RPCs: verb + noun (`CreateVirtualMachine`)
- Messages: PascalCase (`CreateVirtualMachineRequest`)
