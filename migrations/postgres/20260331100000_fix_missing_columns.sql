-- +goose Up
-- +goose StatementBegin

-- Add missing updated_at columns to authz tables
ALTER TABLE role ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE permission ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

-- Add missing priority column to announcements
ALTER TABLE announcements ADD COLUMN IF NOT EXISTS priority INT NOT NULL DEFAULT 0;

-- Grant file.upload to the "user" role
INSERT INTO role_permission (role_id, permission_id)
SELECT r.id, p.id FROM role r, permission p
WHERE r.name = 'user' AND p.name = 'file.upload'
ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE role DROP COLUMN IF EXISTS updated_at;
ALTER TABLE permission DROP COLUMN IF EXISTS updated_at;
ALTER TABLE announcements DROP COLUMN IF EXISTS priority;
-- +goose StatementEnd
