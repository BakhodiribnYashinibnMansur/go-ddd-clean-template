-- +goose Up
ALTER TABLE session ADD COLUMN IF NOT EXISTS device_fingerprint VARCHAR(64);

-- +goose Down
ALTER TABLE session DROP COLUMN IF EXISTS device_fingerprint;
