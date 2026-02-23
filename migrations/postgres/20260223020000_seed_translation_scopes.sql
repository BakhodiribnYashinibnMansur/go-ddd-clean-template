-- +goose Up
-- +goose StatementBegin

-- Add translation scopes
INSERT INTO scope (path, method) VALUES
('/api/v1/translations/*/*', 'GET'),
('/api/v1/translations/*/*', 'PUT'),
('/api/v1/translations/*/*', 'DELETE')
ON CONFLICT DO NOTHING;

-- Add translation permission
INSERT INTO permission (name) VALUES
('translation.manage')
ON CONFLICT DO NOTHING;

-- Link permission to scope
INSERT INTO permission_scope (permission_id, path, method)
SELECT p.id, s.path, s.method FROM permission p, scope s
WHERE p.name = 'translation.manage' AND s.path LIKE '/api/v1/translations%'
ON CONFLICT DO NOTHING;

-- Grant to super_admin (via manager role)
INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r, permission p
WHERE r.name = 'manager' AND p.name = 'translation.manage'
ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM permission_scope WHERE permission_id IN (SELECT id FROM permission WHERE name = 'translation.manage');
DELETE FROM role_permission WHERE permission_id IN (SELECT id FROM permission WHERE name = 'translation.manage');
DELETE FROM permission WHERE name = 'translation.manage';
DELETE FROM scope WHERE path LIKE '/api/v1/translations%';
-- +goose StatementEnd
