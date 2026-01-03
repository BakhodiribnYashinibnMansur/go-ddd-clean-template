-- +goose Up
-- +goose StatementBegin
-- =========================
-- DUMMY DATA SEEDING
-- =========================

-- Roles
INSERT INTO role (name) VALUES ('admin'), ('manager'), ('user');

-- Permissions (Hierarchical sample)
WITH root_perm AS (
    INSERT INTO permission (name) VALUES ('root') RETURNING id
),
sys_perm AS (
    INSERT INTO permission (parent_id, name) 
    SELECT id, 'system' FROM root_perm RETURNING id
),
users_perm AS (
    INSERT INTO permission (parent_id, name)
    SELECT id, 'users' FROM root_perm RETURNING id
)
INSERT INTO permission (parent_id, name)
SELECT id, 'settings' FROM sys_perm;

-- Role Permissions
INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id 
FROM role r, permission p 
WHERE r.name = 'admin' AND p.name = 'root';

-- Users
-- Admin User
INSERT INTO users (role_id, email, phone, username, password_hash, active)
SELECT id, 'admin@system.local', '+998900000000', 'admin', '$2a$10$X7.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1', TRUE
FROM role WHERE name = 'admin';

-- Manager User
INSERT INTO users (role_id, email, phone, username, password_hash, active)
SELECT id, 'manager@system.local', '+998900000001', 'manager', '$2a$10$X7.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1.1', TRUE
FROM role WHERE name = 'manager';

-- Relations
INSERT INTO relation (type, name) VALUES ('REGION', 'Tashkent'), ('BRANCH', 'Chilonzor');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove dummy data
DELETE FROM user_relation;
DELETE FROM role_permission;
DELETE FROM users WHERE email IN ('admin@system.local', 'manager@system.local');
DELETE FROM permission WHERE name IN ('root', 'system', 'users', 'settings');
DELETE FROM role WHERE name IN ('admin', 'manager', 'user');
DELETE FROM relation WHERE name IN ('Tashkent', 'Chilonzor');
-- +goose StatementEnd
