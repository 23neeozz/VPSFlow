CREATE TABLE IF NOT EXISTS agent_sessions (
    session_id TEXT PRIMARY KEY,
    hypervisor_id TEXT NOT NULL UNIQUE,
    agent_version TEXT NOT NULL DEFAULT '',
    node_name TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    last_heartbeat_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS agent_commands (
    command_id TEXT PRIMARY KEY,
    hypervisor_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL DEFAULT '',
    command_type TEXT NOT NULL,
    payload BYTEA NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    result BYTEA,
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_agent_commands_hypervisor_status ON agent_commands(hypervisor_id, status, created_at);
