-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- AUTHZ SEED DATA (Permission, Role, Policy, Scope, Relation)
-- ============================================================================

-- 1. SCOPES (All API endpoints)
INSERT INTO scope (path, method) VALUES 
-- Users
('/api/v1/users', 'GET'),
('/api/v1/users', 'POST'),
('/api/v1/users/*', 'GET'),
('/api/v1/users/*', 'PATCH'),
('/api/v1/users/*', 'DELETE'),
-- Sessions
('/api/v1/users/*/sessions', 'GET'),
('/api/v1/users/*/sessions', 'POST'),
('/api/v1/users/*/sessions', 'DELETE'),
('/api/v1/sessions', 'GET'),
('/api/v1/sessions', 'POST'),
('/api/v1/sessions/revoke-all', 'POST'),
('/api/v1/sessions/current', 'DELETE'),
('/api/v1/sessions/*', 'GET'),
('/api/v1/sessions/*', 'DELETE'),
('/api/v1/sessions/*/activity', 'PUT'),
('/api/v1/sessions/device/*', 'DELETE'),
-- Files (Minio)
('/api/v1/files/upload/images', 'POST'),
('/api/v1/files/upload/image', 'POST'),
('/api/v1/files/upload/doc', 'POST'),
('/api/v1/files/upload/video', 'POST'),
('/api/v1/files/download', 'GET'),
('/api/v1/files/transfer', 'POST'),
-- Authz: Permissions
('/api/v1/authz/permissions', 'POST'),
('/api/v1/authz/permissions', 'GET'),
('/api/v1/authz/permissions/*', 'GET'),
('/api/v1/authz/permissions/*', 'PATCH'),
('/api/v1/authz/permissions/*', 'DELETE'),
('/api/v1/authz/permissions/*/scopes', 'POST'),
('/api/v1/authz/permissions/*/scopes', 'DELETE'),
-- Authz: Roles
('/api/v1/authz/roles', 'POST'),
('/api/v1/authz/roles', 'GET'),
('/api/v1/authz/roles/*', 'GET'),
('/api/v1/authz/roles/*', 'PATCH'),
('/api/v1/authz/roles/*', 'DELETE'),
('/api/v1/authz/roles/*/permissions', 'POST'),
('/api/v1/authz/roles/*/permissions', 'DELETE'),
('/api/v1/authz/roles/*/assign', 'POST'),
-- Authz: Policies
('/api/v1/authz/policies', 'POST'),
('/api/v1/authz/policies', 'GET'),
('/api/v1/authz/policies/*', 'GET'),
('/api/v1/authz/policies/*', 'PATCH'),
('/api/v1/authz/policies/*', 'DELETE'),
-- Authz: Relations
('/api/v1/authz/relations', 'POST'),
('/api/v1/authz/relations', 'GET'),
('/api/v1/authz/relations/*', 'GET'),
('/api/v1/authz/relations/*', 'PATCH'),
('/api/v1/authz/relations/*', 'DELETE'),
('/api/v1/authz/relations/*/users', 'POST'),
('/api/v1/authz/relations/*/users', 'DELETE'),
-- Authz: Scopes
('/api/v1/authz/scopes', 'POST'),
('/api/v1/authz/scopes', 'GET'),
('/api/v1/authz/scopes/*', 'GET'),
('/api/v1/authz/scopes/*', 'DELETE'),
-- Audit & Metrics
('/api/v1/audit/history', 'GET'),
('/api/v1/audit/logs', 'GET'),
('/api/v1/audit/logins', 'GET'),
('/api/v1/audit/sessions', 'GET'),
('/api/v1/audit/actions', 'GET'),
('/api/v1/metrics/functions', 'GET'),
-- Error Codes
('/api/v1/error-codes', 'POST'),
('/api/v1/error-codes', 'GET'),
('/api/v1/error-codes/*', 'GET'),
('/api/v1/error-codes/*', 'PUT'),
-- Integrations
('/api/v1/integrations', 'POST'),
('/api/v1/integrations', 'GET'),
('/api/v1/integrations/*', 'GET'),
('/api/v1/integrations/*', 'PUT'),
('/api/v1/integrations/*', 'DELETE'),
('/api/v1/integrations/*/keys', 'POST'),
('/api/v1/integrations/*/keys', 'GET'),
-- API Keys
('/api/v1/api-keys/*', 'GET'),
('/api/v1/api-keys/*/revoke', 'POST'),
('/api/v1/api-keys/*', 'DELETE'),
-- Feature Flags
('/api/v1/featureflag/boolean', 'GET'),
('/api/v1/featureflag/string', 'GET'),
('/api/v1/featureflag/int', 'GET'),
('/api/v1/featureflag/json', 'GET'),
('/api/v1/featureflag/targeting', 'GET'),
('/api/v1/featureflag/rollout', 'GET')
ON CONFLICT DO NOTHING;

-- 2. PERMISSIONS (Granular)
INSERT INTO permission (name) VALUES 
-- User management
('user.view'), ('user.create'), ('user.update'), ('user.delete'),
-- Session management
('session.view'), ('session.create'), ('session.delete'), ('session.revoke_all'),
-- File management
('file.upload'), ('file.download'), ('file.transfer'),
-- Authz management
('authz.permission.manage'), ('authz.role.manage'), ('authz.policy.manage'), ('authz.relation.manage'), ('authz.scope.manage'),
-- Audit & System
('audit.view'), ('metrics.view'), ('errorcode.manage'),
-- Integrations
('integration.manage'), ('apikey.manage'),
-- Feature Flags
('featureflag.test')
ON CONFLICT DO NOTHING;

-- 3. PERMISSION_SCOPE (Linking)
-- (Simplified for seeding: mapping patterns)

-- User View
INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s 
WHERE p.name = 'user.view' AND s.path LIKE '/api/v1/users%' AND s.method = 'GET'
ON CONFLICT DO NOTHING;

-- User Manage (Create/Update/Delete)
INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s 
WHERE p.name IN ('user.create', 'user.update', 'user.delete') 
AND s.path LIKE '/api/v1/users%' AND s.method IN ('POST', 'PATCH', 'DELETE')
ON CONFLICT DO NOTHING;

-- Session View
INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s
WHERE p.name = 'session.view' AND s.path LIKE '%sessions%' AND s.method = 'GET'
ON CONFLICT DO NOTHING;

-- Session Create
INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s
WHERE p.name = 'session.create' AND s.path LIKE '%sessions%' AND s.method = 'POST'
ON CONFLICT DO NOTHING;

-- Session Delete
INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s
WHERE p.name = 'session.delete' AND s.path LIKE '%sessions%' AND s.method = 'DELETE'
ON CONFLICT DO NOTHING;

-- Session Revoke All
INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s
WHERE p.name = 'session.revoke_all' AND s.path = '/api/v1/sessions/revoke-all' AND s.method = 'POST'
ON CONFLICT DO NOTHING;

-- File Management
INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s 
WHERE p.name = 'file.upload' AND s.path LIKE '%/upload/%' AND s.method = 'POST'
ON CONFLICT DO NOTHING;

INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s 
WHERE p.name = 'file.download' AND s.path = '/api/v1/files/download' AND s.method = 'GET'
ON CONFLICT DO NOTHING;

-- Authz Management (Broad for simplicity in seed)
INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s 
WHERE p.name LIKE 'authz.%' AND s.path LIKE '/api/v1/authz/%'
ON CONFLICT DO NOTHING;

-- Audit
INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s 
WHERE p.name = 'audit.view' AND (s.path LIKE '/api/v1/audit/%' OR s.path LIKE '/api/v1/metrics/%')
ON CONFLICT DO NOTHING;

-- Integrations
INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s 
WHERE p.name IN ('integration.manage', 'apikey.manage') 
AND (s.path LIKE '/api/v1/integrations%' OR s.path LIKE '/api/v1/api-keys%')
ON CONFLICT DO NOTHING;

-- 4. ROLE_PERMISSION (Linking roles to permissions)
-- Manager: Users, Sessions, Files
INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r, permission p 
WHERE r.name = 'manager' AND (p.name LIKE 'user.%' OR p.name LIKE 'session.%' OR p.name LIKE 'file.%')
ON CONFLICT DO NOTHING;

-- User: Manage own profile, own sessions, download files
INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r, permission p
WHERE r.name = 'user' AND p.name IN ('user.view', 'user.update', 'user.delete', 'session.view', 'session.delete', 'session.revoke_all', 'file.download')
ON CONFLICT DO NOTHING;

-- Specialized Roles:
-- Auditor: Read-only access to almost everything
INSERT INTO role (name) VALUES ('auditor') ON CONFLICT DO NOTHING;
INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r, permission p 
WHERE r.name = 'auditor' AND p.name IN ('user.view', 'session.view', 'audit.view', 'metrics.view')
ON CONFLICT DO NOTHING;

-- HR: Manage users and relations
INSERT INTO role (name) VALUES ('hr') ON CONFLICT DO NOTHING;
INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r, permission p 
WHERE r.name = 'hr' AND (p.name LIKE 'user.%' OR p.name = 'authz.relation.manage')
ON CONFLICT DO NOTHING;

-- Support: View users and manage sessions (to assist users)
INSERT INTO role (name) VALUES ('support') ON CONFLICT DO NOTHING;
INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r, permission p 
WHERE r.name = 'support' AND (p.name = 'user.view' OR p.name LIKE 'session.%')
ON CONFLICT DO NOTHING;

-- Developer: Manage integrations and feature flags
INSERT INTO role (name) VALUES ('developer') ON CONFLICT DO NOTHING;
INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r, permission p 
WHERE r.name = 'developer' AND (p.name LIKE 'integration.%' OR p.name LIKE 'apikey.%' OR p.name = 'featureflag.test')
ON CONFLICT DO NOTHING;

-- Viewer: Minimum possible read-only access (Only viewing, no modifications)
INSERT INTO role (name) VALUES ('viewer') ON CONFLICT DO NOTHING;
INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r, permission p 
WHERE r.name = 'viewer' AND p.name IN ('user.view', 'session.view', 'file.download')
ON CONFLICT DO NOTHING;

-- Create a specific restricted user for demonstration
-- Note: Assuming role ID retrieval works via name in a subquery
INSERT INTO users (id, role_id, username, email, phone, password_hash, salt, active, is_approved)
SELECT 
    '00000000-0000-0000-0000-000000000002', 
    (SELECT id FROM role WHERE name = 'viewer'), 
    'viewer_demo', 
    'viewer@example.com', 
    '+998991234567', 
    '$2a$10$x.X/X/X/X/X/X/X/X/X/X.X', -- Dummy hash
    'static_salt', 
    true, 
    true
ON CONFLICT (username) DO NOTHING;

-- 5. RELATIONS (Organizational structure)
INSERT INTO relation (type, name) VALUES 
('REGION', 'Samarkand'), 
('REGION', 'Fergana'),
('BRANCH', 'Yunusobod'),
('BRANCH', 'Mirzo Ulugbek')
ON CONFLICT DO NOTHING;

-- 6. POLICIES (ABAC Examples)
-- Policy: Managers/HR can only manage users if they belong to the same branch.
INSERT INTO policy (permission_id, effect, priority, active)
SELECT p.id, 'ALLOW', 10, true
FROM permission p WHERE p.name IN ('user.update', 'user.delete')
ON CONFLICT DO NOTHING;

INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'policy_conditions', pol.id, 'user.relation_names_any', '$target.user.relation_names', 'string'
FROM policy pol
JOIN permission p ON pol.permission_id = p.id
WHERE p.name IN ('user.update', 'user.delete') AND pol.effect = 'ALLOW' AND pol.priority = 10
ON CONFLICT DO NOTHING;

-- Policy: Auditor can see everything regardless of branch
INSERT INTO policy (permission_id, effect, priority, active)
SELECT p.id, 'ALLOW', 100, true
FROM permission p WHERE p.name = 'user.view'
ON CONFLICT DO NOTHING;

INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'policy_conditions', pol.id, 'user.role_name', 'auditor', 'string'
FROM policy pol
JOIN permission p ON pol.permission_id = p.id
WHERE p.name = 'user.view' AND pol.effect = 'ALLOW' AND pol.priority = 100
ON CONFLICT DO NOTHING;

-- Policy: Restrict deletions to specific IP (example)
INSERT INTO policy (permission_id, effect, priority, active)
SELECT p.id, 'DENY', 100, false
FROM permission p WHERE p.name = 'root'
ON CONFLICT DO NOTHING;

INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'policy_conditions', pol.id, 'env.ip_not_in', '["127.0.0.1", "192.168.1.1"]', 'json_array'
FROM policy pol
JOIN permission p ON pol.permission_id = p.id
WHERE p.name = 'root' AND pol.effect = 'DENY' AND pol.priority = 100
ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM entity_metadata WHERE entity_type = 'policy_conditions';
DELETE FROM role_permission WHERE role_id IN (SELECT id FROM role WHERE name IN ('manager', 'user'));
DELETE FROM permission_scope;
DELETE FROM permission WHERE name IN ('view_users', 'manage_users', 'view_authz', 'manage_authz', 'view_audit');
DELETE FROM relation WHERE name IN ('Samarkand', 'Fergana', 'Yunusobod', 'Mirzo Ulugbek');
DELETE FROM scope WHERE path LIKE '/api/v1/%';
-- +goose StatementEnd
