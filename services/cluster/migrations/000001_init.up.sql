CREATE TABLE IF NOT EXISTS clusters (
    cluster_id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS hypervisors (
    hypervisor_id TEXT PRIMARY KEY,
    cluster_id TEXT REFERENCES clusters(cluster_id) ON DELETE SET NULL,
    node_name TEXT NOT NULL,
    agent_version TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'offline',
    maintenance_mode BOOLEAN NOT NULL DEFAULT FALSE,
    last_heartbeat_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS node_capacity (
    hypervisor_id TEXT PRIMARY KEY REFERENCES hypervisors(hypervisor_id) ON DELETE CASCADE,
    cpu_cores INT NOT NULL DEFAULT 0,
    memory_bytes BIGINT NOT NULL DEFAULT 0,
    storage_bytes BIGINT NOT NULL DEFAULT 0,
    running_vms INT NOT NULL DEFAULT 0,
    allocated_cpu INT NOT NULL DEFAULT 0,
    allocated_memory_bytes BIGINT NOT NULL DEFAULT 0,
    allocated_storage_bytes BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_hypervisors_status ON hypervisors(status);
CREATE INDEX IF NOT EXISTS idx_hypervisors_cluster_id ON hypervisors(cluster_id);

INSERT INTO clusters (cluster_id, name, description)
VALUES ('clu_default', 'default', 'Default compute cluster')
ON CONFLICT (cluster_id) DO NOTHING;
