-- +goose Up
-- +goose StatementBegin
ALTER TABLE session ADD COLUMN revoked INTEGER NOT NULL DEFAULT 0;
CREATE INDEX idx_session_revoked ON session(revoked) WHERE revoked = 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_session_revoked;
ALTER TABLE session DROP COLUMN revoked;
-- +goose StatementEnd
