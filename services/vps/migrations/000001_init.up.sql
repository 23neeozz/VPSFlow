CREATE TABLE IF NOT EXISTS vps_instances (
    vps_id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    vm_id TEXT NOT NULL,
    name TEXT NOT NULL,
    owner_user_id TEXT NOT NULL,
    owner_email TEXT NOT NULL DEFAULT '',
    assigned_by_user_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE (tenant_id, name)
);

CREATE INDEX IF NOT EXISTS idx_vps_tenant ON vps_instances(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vps_owner ON vps_instances(tenant_id, owner_user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vps_vm ON vps_instances(vm_id) WHERE deleted_at IS NULL;
