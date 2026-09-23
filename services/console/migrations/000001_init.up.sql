CREATE TABLE IF NOT EXISTS console_sessions (
    session_id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    vm_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'active',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_console_sessions_vm ON console_sessions(vm_id);
