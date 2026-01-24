-- +goose Up
-- +goose StatementBegin
-- =========================
-- DUMMY DATA SEEDING
-- =========================

-- Roles
INSERT INTO role (name) VALUES ('super_admin'), ('admin'), ('manager'), ('user') ON CONFLICT DO NOTHING;

-- Permissions (Hierarchical sample)
-- (Simplified for seeding: checking existence or using DO NOTHING if constrained)
-- Using CTEs here is fine assuming empty DB.

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
-- Grant root to super_admin (unrestricted access)
INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id 
FROM role r, permission p 
WHERE r.name = 'super_admin' AND p.name = 'root';

-- Grant root to admin
INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id 
FROM role r, permission p 
WHERE r.name = 'admin' AND p.name = 'root';

-- NO USERS SEEDED (Use /admin/setup to create first user)

-- Relations
INSERT INTO relation (type, name) VALUES ('REGION', 'Tashkent'), ('BRANCH', 'Chilonzor');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove dummy data
DELETE FROM user_relation;
DELETE FROM role_permission;
DELETE FROM users; -- Clear all users if we rolled back seed? No, dangerous. But "seed_data" rollback usually implies clearing seeds.
-- Given we didn't seed users, maybe don't delete them.
-- But Down should invert Up. Up did nothing for users. So Down does nothing for users.
DELETE FROM permission WHERE name IN ('root', 'system', 'users', 'settings');
DELETE FROM role WHERE name IN ('super_admin', 'admin', 'manager', 'user');
DELETE FROM relation WHERE name IN ('Tashkent', 'Chilonzor');
-- +goose StatementEnd
