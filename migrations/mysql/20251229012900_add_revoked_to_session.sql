-- +goose Up
-- +goose StatementBegin
ALTER TABLE session 
ADD COLUMN revoked BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX idx_session_revoked ON session(revoked);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_session_revoked ON session;
ALTER TABLE session DROP COLUMN revoked;
-- +goose StatementEnd
