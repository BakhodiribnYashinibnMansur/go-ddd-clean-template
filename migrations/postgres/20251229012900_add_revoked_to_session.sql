-- +goose Up
-- +goose StatementBegin
ALTER TABLE session ADD COLUMN IF NOT EXISTS revoked BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX IF NOT EXISTS idx_session_revoked ON session(revoked) WHERE revoked = FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_session_revoked;
ALTER TABLE session DROP COLUMN IF EXISTS revoked;
-- +goose StatementEnd
