-- +goose Up
ALTER TABLE session ADD COLUMN IF NOT EXISTS previous_refresh_hash VARCHAR(512);

-- +goose Down
ALTER TABLE session DROP COLUMN IF EXISTS previous_refresh_hash;
