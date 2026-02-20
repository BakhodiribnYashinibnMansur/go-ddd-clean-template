-- +goose Up
-- +goose StatementBegin
ALTER TABLE policy ADD COLUMN IF NOT EXISTS conditions JSONB DEFAULT '{}';
ALTER TABLE users ADD COLUMN IF NOT EXISTS attributes JSONB NOT NULL DEFAULT '{}';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Rollback not safe if columns were already expected to be there
-- +goose StatementEnd
