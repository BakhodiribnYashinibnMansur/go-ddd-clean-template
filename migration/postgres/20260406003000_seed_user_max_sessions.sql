-- +goose Up
-- +goose StatementBegin
INSERT INTO site_settings (key, value, value_type, category, description, is_public)
VALUES (
    'user.max_sessions',
    '3',
    'integer',
    'general',
    'Maximum number of concurrent active sessions per user across all integrations',
    false
)
ON CONFLICT (key) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM site_settings WHERE key = 'user.max_sessions';
-- +goose StatementEnd
