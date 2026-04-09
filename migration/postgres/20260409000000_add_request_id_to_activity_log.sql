-- +goose Up
-- +goose StatementBegin
ALTER TABLE activity_log ADD COLUMN request_id VARCHAR(64);
CREATE INDEX IF NOT EXISTS idx_activity_log_request_id ON activity_log (request_id) WHERE request_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_activity_log_request_id;
ALTER TABLE activity_log DROP COLUMN IF EXISTS request_id;
-- +goose StatementEnd
