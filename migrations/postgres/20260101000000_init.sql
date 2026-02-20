-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- EXTENSIONS
-- ============================================================================
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- ENUM TYPES
-- ============================================================================
CREATE TYPE relation_types AS ENUM ('UNREVEALED', 'BRANCH', 'REGION');
CREATE TYPE policy_effect AS ENUM ('ALLOW', 'DENY');
CREATE TYPE session_device_type AS ENUM ('DESKTOP', 'MOBILE', 'TABLET', 'BOT', 'TV');
CREATE TYPE audit_action_type AS ENUM (
    'LOGIN', 'LOGOUT', 'SESSION_REVOKE',
    'PASSWORD_CHANGE', 'MFA_VERIFY_FAIL',
    'ACCESS_GRANTED', 'ACCESS_DENIED',
    'POLICY_MATCHED', 'POLICY_DENIED',
    'USER_CREATE', 'USER_UPDATE', 'USER_DELETE',
    'ROLE_ASSIGN', 'ROLE_REMOVE',
    'ORDER_APPROVE', 'ORDER_CANCEL',
    'PAYMENT_PROCESS', 'PAYMENT_CANCEL',
    'POLICY_EVALUATED',
    'ADMIN_CHANGE'
);
CREATE TYPE error_severity_enum AS ENUM ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL');
CREATE TYPE error_category_enum AS ENUM ('DATA', 'AUTH', 'SYSTEM', 'VALIDATION', 'BUSINESS', 'UNKNOWN');

-- ============================================================================
-- TABLES
-- ============================================================================

-- RBAC: Role & Permission
CREATE TABLE role (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE permission (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID REFERENCES permission(id),
    name VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (parent_id, name)
);

CREATE TABLE role_permission (
    role_id UUID REFERENCES role(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permission(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, permission_id)
);

-- API Scope
CREATE TABLE scope (
    path VARCHAR NOT NULL,
    method VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (path, method)
);

CREATE TABLE permission_scope (
    permission_id UUID REFERENCES permission(id) ON DELETE CASCADE,
    path VARCHAR,
    method VARCHAR,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (permission_id, path, method),
    FOREIGN KEY (path, method) REFERENCES scope(path, method)
);

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID REFERENCES role(id),
    username VARCHAR UNIQUE,
    email VARCHAR UNIQUE,
    phone VARCHAR UNIQUE,
    password_hash TEXT,
    salt VARCHAR,
    attributes JSONB NOT NULL DEFAULT '{}',
    active BOOLEAN DEFAULT TRUE,
    is_approved BOOLEAN DEFAULT FALSE,
    last_seen TIMESTAMP,
    deleted_at BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- Session
CREATE TABLE session (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id UUID NOT NULL,
    device_name VARCHAR(255),
    device_type session_device_type,
    ip_address INET,
    user_agent VARCHAR(512),
    fcm_token VARCHAR(512),
    data JSONB,
    refresh_token_hash VARCHAR(512),
    expires_at TIMESTAMP NOT NULL,
    last_activity TIMESTAMP NOT NULL DEFAULT NOW(),
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    os VARCHAR(100),
    os_version VARCHAR(50),
    browser VARCHAR(100),
    browser_version VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_session_user_id ON session(user_id);
CREATE INDEX idx_session_device_id ON session(device_id);
CREATE INDEX idx_session_expires_at ON session(expires_at);
CREATE INDEX idx_session_last_activity ON session(last_activity);
CREATE INDEX idx_session_revoked ON session(revoked) WHERE revoked = FALSE;
CREATE INDEX idx_session_os ON session(os);
CREATE INDEX idx_session_browser ON session(browser);

-- Relation (Org Structure)
CREATE TABLE relation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type relation_types NOT NULL,
    name VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (type, name)
);

CREATE TABLE user_relation (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    relation_id UUID REFERENCES relation(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, relation_id)
);

-- Policy (ABAC)
CREATE TABLE policy (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    permission_id UUID REFERENCES permission(id) ON DELETE CASCADE,
    effect policy_effect NOT NULL,
    priority INT DEFAULT 100,
    active BOOLEAN DEFAULT TRUE,
    conditions JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_policy_permission_id ON policy(permission_id);

-- Audit Log
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    session_id UUID REFERENCES session(id),
    action audit_action_type NOT NULL,
    resource_type VARCHAR(64),
    resource_id UUID,
    platform VARCHAR(16),
    ip_address INET,
    user_agent VARCHAR(512),
    permission VARCHAR(128),
    policy_id UUID REFERENCES policy(id),
    decision VARCHAR(16),
    success BOOLEAN NOT NULL,
    error_message TEXT,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_user_id ON audit_log(user_id);
CREATE INDEX idx_audit_session_id ON audit_log(session_id);
CREATE INDEX idx_audit_action ON audit_log(action);
CREATE INDEX idx_audit_resource ON audit_log(resource_type, resource_id);
CREATE INDEX idx_audit_created_at ON audit_log(created_at);
CREATE INDEX idx_audit_decision ON audit_log(decision) WHERE decision IS NOT NULL;
CREATE INDEX idx_audit_policy_id ON audit_log(policy_id) WHERE policy_id IS NOT NULL;
CREATE INDEX idx_audit_failed_attempts ON audit_log(created_at, user_id, action) WHERE success = FALSE;

-- Endpoint History
CREATE TABLE endpoint_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    session_id UUID REFERENCES session(id),
    method VARCHAR(8) NOT NULL,
    path VARCHAR(255) NOT NULL,
    status_code SMALLINT NOT NULL,
    duration_ms INTEGER NOT NULL,
    platform VARCHAR(16),
    ip_address INET,
    user_agent VARCHAR(512),
    permission VARCHAR(128),
    decision VARCHAR(16),
    request_id UUID,
    rate_limited BOOLEAN DEFAULT FALSE,
    response_size INTEGER,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_eh_user_id ON endpoint_history(user_id);
CREATE INDEX idx_eh_session_id ON endpoint_history(session_id);
CREATE INDEX idx_eh_path ON endpoint_history(path);
CREATE INDEX idx_eh_method ON endpoint_history(method);
CREATE INDEX idx_eh_status ON endpoint_history(status_code);
CREATE INDEX idx_eh_created_at ON endpoint_history(created_at);
CREATE INDEX idx_eh_user_created ON endpoint_history(user_id, created_at);
CREATE INDEX idx_eh_path_status ON endpoint_history(path, status_code);
CREATE INDEX idx_eh_decision ON endpoint_history(decision) WHERE decision IS NOT NULL;
CREATE INDEX idx_eh_errors ON endpoint_history(created_at, path, status_code) WHERE status_code >= 500;
CREATE INDEX idx_eh_slow_requests ON endpoint_history(created_at, path, duration_ms) WHERE duration_ms > 1000;

-- System Errors
CREATE TABLE system_errors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(64) NOT NULL,
    message TEXT NOT NULL,
    stack_trace TEXT,
    metadata JSONB,
    severity VARCHAR(16) NOT NULL DEFAULT 'ERROR',
    service_name VARCHAR(64) DEFAULT 'api',
    request_id UUID,
    user_id UUID,
    ip_address INET,
    path VARCHAR(255),
    method VARCHAR(8),
    is_resolved BOOLEAN DEFAULT FALSE,
    resolved_at TIMESTAMP,
    resolved_by UUID,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sys_err_code ON system_errors(code);
CREATE INDEX idx_sys_err_severity ON system_errors(severity);
CREATE INDEX idx_sys_err_created_at ON system_errors(created_at);
CREATE INDEX idx_sys_err_req_id ON system_errors(request_id) WHERE request_id IS NOT NULL;
CREATE INDEX idx_sys_err_resolved ON system_errors(is_resolved);

-- Function Metrics
CREATE TABLE function_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    latency_ms INTEGER NOT NULL,
    is_panic BOOLEAN DEFAULT FALSE,
    panic_error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_func_metrics_name ON function_metrics(name);
CREATE INDEX idx_func_metrics_created_at ON function_metrics(created_at);
CREATE INDEX idx_func_metrics_panic ON function_metrics(is_panic);

-- Site Settings
CREATE TABLE site_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(64) UNIQUE NOT NULL,
    value TEXT,
    value_type VARCHAR(16) NOT NULL DEFAULT 'string',
    category VARCHAR(32) NOT NULL DEFAULT 'general',
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_site_settings_key ON site_settings(key);
CREATE INDEX idx_site_settings_category ON site_settings(category);
CREATE INDEX idx_site_settings_public ON site_settings(is_public);

-- Error Codes
CREATE TABLE error_code (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(255) NOT NULL UNIQUE,
    message TEXT NOT NULL,
    http_status INT NOT NULL,
    category error_category_enum DEFAULT 'UNKNOWN',
    severity error_severity_enum DEFAULT 'MEDIUM',
    retryable BOOLEAN DEFAULT FALSE,
    retry_after INT DEFAULT 0,
    suggestion TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_error_code_code ON error_code(code);
CREATE INDEX idx_error_code_category ON error_code(category);

-- Integrations
CREATE TABLE integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    base_url VARCHAR(500) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    config JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_id UUID NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    key VARCHAR(255) NOT NULL UNIQUE,
    key_prefix VARCHAR(20) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_integrations_name ON integrations(name) WHERE deleted_at IS NULL;
CREATE INDEX idx_integrations_is_active ON integrations(is_active) WHERE deleted_at IS NULL;
CREATE INDEX idx_api_keys_integration_id ON api_keys(integration_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_api_keys_key ON api_keys(key) WHERE deleted_at IS NULL AND is_active = true;
CREATE INDEX idx_api_keys_key_prefix ON api_keys(key_prefix) WHERE deleted_at IS NULL;

COMMENT ON TABLE integrations IS 'Stores third-party integration platform configurations';
COMMENT ON TABLE api_keys IS 'Stores API keys for accessing integrations';
COMMENT ON COLUMN api_keys.key IS 'Hashed API key for security';
COMMENT ON COLUMN api_keys.key_prefix IS 'Visible prefix for key identification (e.g., sk_live_abc...)';

-- ============================================================================
-- TRIGGER FUNCTIONS
-- ============================================================================
CREATE OR REPLACE FUNCTION notify_cache_invalidation() RETURNS TRIGGER AS $$
DECLARE
    payload TEXT;
BEGIN
    payload := TG_TABLE_NAME;
    PERFORM pg_notify('cache_invalidation', payload);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_integrations_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_api_keys_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Cache invalidation
CREATE TRIGGER invalidate_cache_users AFTER INSERT OR UPDATE OR DELETE ON users FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();
CREATE TRIGGER invalidate_cache_role AFTER INSERT OR UPDATE OR DELETE ON role FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();
CREATE TRIGGER invalidate_cache_permission AFTER INSERT OR UPDATE OR DELETE ON permission FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();
CREATE TRIGGER invalidate_cache_policy AFTER INSERT OR UPDATE OR DELETE ON policy FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();
CREATE TRIGGER invalidate_cache_session AFTER INSERT OR UPDATE OR DELETE ON session FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();
CREATE TRIGGER invalidate_cache_relation AFTER INSERT OR UPDATE OR DELETE ON relation FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();
CREATE TRIGGER invalidate_cache_integrations AFTER INSERT OR UPDATE OR DELETE ON integrations FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();
CREATE TRIGGER invalidate_cache_api_keys AFTER INSERT OR UPDATE OR DELETE ON api_keys FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

-- Updated_at triggers
CREATE TRIGGER trigger_integrations_updated_at BEFORE UPDATE ON integrations FOR EACH ROW EXECUTE FUNCTION update_integrations_updated_at();
CREATE TRIGGER trigger_api_keys_updated_at BEFORE UPDATE ON api_keys FOR EACH ROW EXECUTE FUNCTION update_api_keys_updated_at();

-- ============================================================================
-- SEED DATA
-- ============================================================================

-- Roles
INSERT INTO role (name) VALUES ('super_admin'), ('admin'), ('manager'), ('user') ON CONFLICT DO NOTHING;

-- Permissions (Hierarchical)
WITH root_perm AS (
    INSERT INTO permission (name) VALUES ('root') RETURNING id
),
sys_perm AS (
    INSERT INTO permission (parent_id, name) SELECT id, 'system' FROM root_perm RETURNING id
),
users_perm AS (
    INSERT INTO permission (parent_id, name) SELECT id, 'users' FROM root_perm RETURNING id
)
INSERT INTO permission (parent_id, name) SELECT id, 'settings' FROM sys_perm;

-- Role Permissions (root -> super_admin, admin)
INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r, permission p WHERE r.name = 'super_admin' AND p.name = 'root';

INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r, permission p WHERE r.name = 'admin' AND p.name = 'root';

-- Relations
INSERT INTO relation (type, name) VALUES ('REGION', 'Tashkent'), ('BRANCH', 'Chilonzor');

-- Site Settings
INSERT INTO site_settings (key, value, value_type, category, description, is_public) VALUES
    ('site_name', 'Go Clean Template', 'string', 'general', 'Application name', true),
    ('site_description', 'A clean architecture template for Go applications', 'string', 'general', 'Site description', true),
    ('maintenance_mode', 'false', 'boolean', 'maintenance', 'Enable maintenance mode', false),
    ('maintenance_message', 'We are currently performing maintenance. Please check back soon.', 'string', 'maintenance', 'Maintenance mode message', true),
    ('allow_registration', 'true', 'boolean', 'general', 'Allow new user registration', false),
    ('max_upload_size', '10485760', 'integer', 'general', 'Maximum upload size in bytes (10MB)', false),
    ('session_timeout', '3600', 'integer', 'general', 'Session timeout in seconds', false),
    ('admin_email', 'admin@example.com', 'string', 'email', 'Admin email address', false),
    ('smtp_enabled', 'false', 'boolean', 'email', 'Enable SMTP email sending', false),
    ('items_per_page', '10', 'integer', 'general', 'Default items per page', false);

-- Error Code seed
INSERT INTO error_code (code, message, http_status, category, severity, retryable, retry_after, suggestion)
VALUES ('RESOURCE_NOT_FOUND', 'The requested resource was not found.', 404, 'DATA', 'MEDIUM', FALSE, 5, 'Please check our documentation.');

-- CI Admin user
INSERT INTO users (id, role_id, username, email, phone, password_hash, active, is_approved)
SELECT '00000000-0000-0000-0000-000000000001', id, 'admin', 'admin@test.com', '+998901234567',
    '$2a$10$vI8aWBnW3fID.97.kHjSLe9M8U.RE9C7kY1R.9WJ.H.9WJ.H.9WJ.H', true, true
FROM role WHERE name = 'admin'
ON CONFLICT DO NOTHING;

-- ============================================================================
-- AUTHZ SEED DATA
-- ============================================================================

-- Scopes (All API endpoints)
INSERT INTO scope (path, method) VALUES
('/api/v1/users', 'GET'), ('/api/v1/users', 'POST'),
('/api/v1/users/*', 'GET'), ('/api/v1/users/*', 'PATCH'), ('/api/v1/users/*', 'DELETE'),
('/api/v1/users/*/sessions', 'GET'), ('/api/v1/users/*/sessions', 'POST'), ('/api/v1/users/*/sessions', 'DELETE'),
('/api/v1/sessions', 'GET'), ('/api/v1/sessions', 'POST'),
('/api/v1/sessions/revoke-all', 'POST'), ('/api/v1/sessions/current', 'DELETE'),
('/api/v1/sessions/*', 'GET'), ('/api/v1/sessions/*', 'DELETE'),
('/api/v1/sessions/*/activity', 'PUT'), ('/api/v1/sessions/device/*', 'DELETE'),
('/api/v1/files/upload/images', 'POST'), ('/api/v1/files/upload/image', 'POST'),
('/api/v1/files/upload/doc', 'POST'), ('/api/v1/files/upload/video', 'POST'),
('/api/v1/files/download', 'GET'), ('/api/v1/files/transfer', 'POST'),
('/api/v1/authz/permissions', 'POST'), ('/api/v1/authz/permissions', 'GET'),
('/api/v1/authz/permissions/*', 'GET'), ('/api/v1/authz/permissions/*', 'PATCH'), ('/api/v1/authz/permissions/*', 'DELETE'),
('/api/v1/authz/permissions/*/scopes', 'POST'), ('/api/v1/authz/permissions/*/scopes', 'DELETE'),
('/api/v1/authz/roles', 'POST'), ('/api/v1/authz/roles', 'GET'),
('/api/v1/authz/roles/*', 'GET'), ('/api/v1/authz/roles/*', 'PATCH'), ('/api/v1/authz/roles/*', 'DELETE'),
('/api/v1/authz/roles/*/permissions', 'POST'), ('/api/v1/authz/roles/*/permissions', 'DELETE'),
('/api/v1/authz/roles/*/assign', 'POST'),
('/api/v1/authz/policies', 'POST'), ('/api/v1/authz/policies', 'GET'),
('/api/v1/authz/policies/*', 'GET'), ('/api/v1/authz/policies/*', 'PATCH'), ('/api/v1/authz/policies/*', 'DELETE'),
('/api/v1/authz/relations', 'POST'), ('/api/v1/authz/relations', 'GET'),
('/api/v1/authz/relations/*', 'GET'), ('/api/v1/authz/relations/*', 'PATCH'), ('/api/v1/authz/relations/*', 'DELETE'),
('/api/v1/authz/relations/*/users', 'POST'), ('/api/v1/authz/relations/*/users', 'DELETE'),
('/api/v1/authz/scopes', 'POST'), ('/api/v1/authz/scopes', 'GET'),
('/api/v1/authz/scopes/*', 'GET'), ('/api/v1/authz/scopes/*', 'DELETE'),
('/api/v1/audit/history', 'GET'), ('/api/v1/audit/logs', 'GET'),
('/api/v1/audit/logins', 'GET'), ('/api/v1/audit/sessions', 'GET'), ('/api/v1/audit/actions', 'GET'),
('/api/v1/metrics/functions', 'GET'),
('/api/v1/error-codes', 'POST'), ('/api/v1/error-codes', 'GET'),
('/api/v1/error-codes/*', 'GET'), ('/api/v1/error-codes/*', 'PUT'),
('/api/v1/integrations', 'POST'), ('/api/v1/integrations', 'GET'),
('/api/v1/integrations/*', 'GET'), ('/api/v1/integrations/*', 'PUT'), ('/api/v1/integrations/*', 'DELETE'),
('/api/v1/integrations/*/keys', 'POST'), ('/api/v1/integrations/*/keys', 'GET'),
('/api/v1/api-keys/*', 'GET'), ('/api/v1/api-keys/*/revoke', 'POST'), ('/api/v1/api-keys/*', 'DELETE'),
('/api/v1/featureflag/boolean', 'GET'), ('/api/v1/featureflag/string', 'GET'),
('/api/v1/featureflag/int', 'GET'), ('/api/v1/featureflag/json', 'GET'),
('/api/v1/featureflag/targeting', 'GET'), ('/api/v1/featureflag/rollout', 'GET')
ON CONFLICT DO NOTHING;

-- Granular permissions
INSERT INTO permission (name) VALUES
('user.view'), ('user.create'), ('user.update'), ('user.delete'),
('session.view'), ('session.create'), ('session.delete'), ('session.revoke_all'),
('file.upload'), ('file.download'), ('file.transfer'),
('authz.permission.manage'), ('authz.role.manage'), ('authz.policy.manage'), ('authz.relation.manage'), ('authz.scope.manage'),
('audit.view'), ('metrics.view'), ('errorcode.manage'),
('integration.manage'), ('apikey.manage'),
('featureflag.test')
ON CONFLICT DO NOTHING;

-- Permission-Scope linking
INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s
WHERE p.name = 'user.view' AND s.path LIKE '/api/v1/users%' AND s.method = 'GET'
ON CONFLICT DO NOTHING;

INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s
WHERE p.name IN ('user.create', 'user.update', 'user.delete')
AND s.path LIKE '/api/v1/users%' AND s.method IN ('POST', 'PATCH', 'DELETE')
ON CONFLICT DO NOTHING;

INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s
WHERE p.name = 'session.view' AND s.path LIKE '%sessions%' AND s.method = 'GET'
ON CONFLICT DO NOTHING;

INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s
WHERE p.name = 'session.create' AND s.path LIKE '%sessions%' AND s.method = 'POST'
ON CONFLICT DO NOTHING;

INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s
WHERE p.name = 'session.delete' AND s.path LIKE '%sessions%' AND s.method = 'DELETE'
ON CONFLICT DO NOTHING;

INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s
WHERE p.name = 'session.revoke_all' AND s.path = '/api/v1/sessions/revoke-all' AND s.method = 'POST'
ON CONFLICT DO NOTHING;

INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s
WHERE p.name = 'file.upload' AND s.path LIKE '%/upload/%' AND s.method = 'POST'
ON CONFLICT DO NOTHING;

INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s
WHERE p.name = 'file.download' AND s.path = '/api/v1/files/download' AND s.method = 'GET'
ON CONFLICT DO NOTHING;

INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s
WHERE p.name LIKE 'authz.%' AND s.path LIKE '/api/v1/authz/%'
ON CONFLICT DO NOTHING;

INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s
WHERE p.name = 'audit.view' AND (s.path LIKE '/api/v1/audit/%' OR s.path LIKE '/api/v1/metrics/%')
ON CONFLICT DO NOTHING;

INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s
WHERE p.name IN ('integration.manage', 'apikey.manage')
AND (s.path LIKE '/api/v1/integrations%' OR s.path LIKE '/api/v1/api-keys%')
ON CONFLICT DO NOTHING;

-- Role-Permission linking
INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r, permission p
WHERE r.name = 'manager' AND (p.name LIKE 'user.%' OR p.name LIKE 'session.%' OR p.name LIKE 'file.%')
ON CONFLICT DO NOTHING;

INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r, permission p
WHERE r.name = 'user' AND p.name IN ('user.view', 'user.update', 'user.delete', 'session.view', 'session.delete', 'session.revoke_all', 'file.download')
ON CONFLICT DO NOTHING;

-- Specialized roles
INSERT INTO role (name) VALUES ('auditor'), ('hr'), ('support'), ('developer'), ('viewer') ON CONFLICT DO NOTHING;

INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r, permission p
WHERE r.name = 'auditor' AND p.name IN ('user.view', 'session.view', 'audit.view', 'metrics.view')
ON CONFLICT DO NOTHING;

INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r, permission p
WHERE r.name = 'hr' AND (p.name LIKE 'user.%' OR p.name = 'authz.relation.manage')
ON CONFLICT DO NOTHING;

INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r, permission p
WHERE r.name = 'support' AND (p.name = 'user.view' OR p.name LIKE 'session.%')
ON CONFLICT DO NOTHING;

INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r, permission p
WHERE r.name = 'developer' AND (p.name LIKE 'integration.%' OR p.name LIKE 'apikey.%' OR p.name = 'featureflag.test')
ON CONFLICT DO NOTHING;

INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r, permission p
WHERE r.name = 'viewer' AND p.name IN ('user.view', 'session.view', 'file.download')
ON CONFLICT DO NOTHING;

-- Demo viewer user
INSERT INTO users (id, role_id, username, email, phone, password_hash, salt, active, is_approved)
SELECT '00000000-0000-0000-0000-000000000002', (SELECT id FROM role WHERE name = 'viewer'),
    'viewer_demo', 'viewer@example.com', '+998991234567', '$2a$10$x.X/X/X/X/X/X/X/X/X/X.X', 'static_salt', true, true
ON CONFLICT (username) DO NOTHING;

-- Additional relations
INSERT INTO relation (type, name) VALUES
('REGION', 'Samarkand'), ('REGION', 'Fergana'),
('BRANCH', 'Yunusobod'), ('BRANCH', 'Mirzo Ulugbek')
ON CONFLICT DO NOTHING;

-- ABAC Policies
INSERT INTO policy (permission_id, effect, priority, active, conditions)
SELECT p.id, 'ALLOW', 10, true,
'{"user.relation_names_any": "$target.user.relation_names"}'::jsonb
FROM permission p WHERE p.name IN ('user.update', 'user.delete')
ON CONFLICT DO NOTHING;

INSERT INTO policy (permission_id, effect, priority, active, conditions)
SELECT p.id, 'ALLOW', 100, true,
'{"user.role_name": "auditor"}'::jsonb
FROM permission p WHERE p.name = 'user.view'
ON CONFLICT DO NOTHING;

INSERT INTO policy (permission_id, effect, priority, active, conditions)
SELECT p.id, 'DENY', 100, false,
'{"env.ip_not_in": ["127.0.0.1", "192.168.1.1"]}'::jsonb
FROM permission p WHERE p.name = 'root'
ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trigger_api_keys_updated_at ON api_keys;
DROP TRIGGER IF EXISTS trigger_integrations_updated_at ON integrations;
DROP TRIGGER IF EXISTS invalidate_cache_api_keys ON api_keys;
DROP TRIGGER IF EXISTS invalidate_cache_integrations ON integrations;
DROP TRIGGER IF EXISTS invalidate_cache_relation ON relation;
DROP TRIGGER IF EXISTS invalidate_cache_session ON session;
DROP TRIGGER IF EXISTS invalidate_cache_policy ON policy;
DROP TRIGGER IF EXISTS invalidate_cache_permission ON permission;
DROP TRIGGER IF EXISTS invalidate_cache_role ON role;
DROP TRIGGER IF EXISTS invalidate_cache_users ON users;

DROP TABLE IF EXISTS api_keys CASCADE;
DROP TABLE IF EXISTS integrations CASCADE;
DROP TABLE IF EXISTS error_code CASCADE;
DROP TABLE IF EXISTS site_settings CASCADE;
DROP TABLE IF EXISTS function_metrics CASCADE;
DROP TABLE IF EXISTS system_errors CASCADE;
DROP TABLE IF EXISTS endpoint_history CASCADE;
DROP TABLE IF EXISTS audit_log CASCADE;
DROP TABLE IF EXISTS policy CASCADE;
DROP TABLE IF EXISTS user_relation CASCADE;
DROP TABLE IF EXISTS relation CASCADE;
DROP TABLE IF EXISTS session CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS permission_scope CASCADE;
DROP TABLE IF EXISTS scope CASCADE;
DROP TABLE IF EXISTS role_permission CASCADE;
DROP TABLE IF EXISTS permission CASCADE;
DROP TABLE IF EXISTS role CASCADE;

DROP FUNCTION IF EXISTS update_api_keys_updated_at() CASCADE;
DROP FUNCTION IF EXISTS update_integrations_updated_at() CASCADE;
DROP FUNCTION IF EXISTS notify_cache_invalidation() CASCADE;

DROP TYPE IF EXISTS error_category_enum;
DROP TYPE IF EXISTS error_severity_enum;
DROP TYPE IF EXISTS audit_action_type;
DROP TYPE IF EXISTS session_device_type;
DROP TYPE IF EXISTS policy_effect;
DROP TYPE IF EXISTS relation_types;
-- +goose StatementEnd
