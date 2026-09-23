CREATE TABLE IF NOT EXISTS permissions (
    name TEXT PRIMARY KEY,
    description TEXT NOT NULL,
    resource TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS roles (
    role_id TEXT PRIMARY KEY,
    tenant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, name)
);

CREATE INDEX IF NOT EXISTS idx_roles_tenant_id ON roles(tenant_id);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id TEXT NOT NULL REFERENCES roles(role_id) ON DELETE CASCADE,
    permission_name TEXT NOT NULL REFERENCES permissions(name),
    PRIMARY KEY (role_id, permission_name)
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    role_id TEXT NOT NULL REFERENCES roles(role_id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, tenant_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_user_roles_tenant_user ON user_roles(tenant_id, user_id);

INSERT INTO permissions (name, description, resource) VALUES
    ('vm.create', 'Create virtual machines', 'vm'),
    ('vm.read', 'View virtual machines', 'vm'),
    ('vm.update', 'Update virtual machines', 'vm'),
    ('vm.delete', 'Delete virtual machines', 'vm'),
    ('vm.console', 'Access VM console', 'vm'),
    ('network.create', 'Create networks', 'network'),
    ('network.read', 'View networks', 'network'),
    ('network.update', 'Update networks', 'network'),
    ('network.delete', 'Delete networks', 'network'),
    ('storage.create', 'Create storage volumes', 'storage'),
    ('storage.read', 'View storage volumes', 'storage'),
    ('storage.update', 'Update storage volumes', 'storage'),
    ('storage.delete', 'Delete storage volumes', 'storage'),
    ('tenant.manage', 'Manage tenant settings', 'tenant'),
    ('tenant.billing', 'Manage tenant billing', 'tenant'),
    ('tenant.members', 'Manage tenant members', 'tenant'),
    ('iam.roles.manage', 'Manage IAM roles', 'iam'),
    ('iam.users.manage', 'Manage IAM user assignments', 'iam'),
    ('audit.read', 'Read audit logs', 'audit'),
    ('admin.hypervisors.manage', 'Manage hypervisors', 'admin')
ON CONFLICT (name) DO NOTHING;
