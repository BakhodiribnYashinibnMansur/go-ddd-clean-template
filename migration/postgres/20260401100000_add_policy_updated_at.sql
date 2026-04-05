-- +goose Up
-- +goose StatementBegin
ALTER TABLE policy ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE policy DROP COLUMN IF EXISTS updated_at;
-- +goose StatementEnd
