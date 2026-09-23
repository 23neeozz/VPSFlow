CREATE TABLE IF NOT EXISTS virtual_machines (
    vm_id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    project_id TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    hypervisor_id TEXT,
    domain_name TEXT,
    vcpus INT NOT NULL,
    memory_mb INT NOT NULL,
    disk_gb INT NOT NULL,
    image_ref TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (tenant_id, name)
);

CREATE TABLE IF NOT EXISTS vm_operations (
    operation_id TEXT PRIMARY KEY,
    vm_id TEXT NOT NULL REFERENCES virtual_machines(vm_id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    operation_type TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    command_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS idx_virtual_machines_tenant ON virtual_machines(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vm_operations_vm_id ON vm_operations(vm_id);
