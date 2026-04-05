-- +goose Up
-- +goose StatementBegin
ALTER TABLE session
    ADD COLUMN integration_name TEXT NOT NULL DEFAULT 'gct-client';
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_session_user_id_integration_name
    ON session (user_id, integration_name)
    WHERE revoked = false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_session_user_id_integration_name;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE session DROP COLUMN IF EXISTS integration_name;
-- +goose StatementEnd
