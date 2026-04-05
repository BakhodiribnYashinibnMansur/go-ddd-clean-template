-- +goose Up
-- +goose StatementBegin
ALTER TABLE integrations
    ADD COLUMN jwt_api_key_hash BYTEA,
    ADD COLUMN jwt_access_ttl_seconds INTEGER,
    ADD COLUMN jwt_refresh_ttl_seconds INTEGER,
    ADD COLUMN jwt_public_key_pem TEXT,
    ADD COLUMN jwt_previous_public_key_pem TEXT,
    ADD COLUMN jwt_key_id TEXT,
    ADD COLUMN jwt_previous_key_id TEXT,
    ADD COLUMN jwt_rotated_at TIMESTAMPTZ,
    ADD COLUMN jwt_rotate_every_days INTEGER NOT NULL DEFAULT 30,
    ADD COLUMN jwt_binding_mode TEXT NOT NULL DEFAULT 'warn',
    ADD COLUMN jwt_max_sessions INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE UNIQUE INDEX idx_integrations_jwt_api_key_hash
    ON integrations (jwt_api_key_hash)
    WHERE jwt_api_key_hash IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_integrations_jwt_api_key_hash;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE integrations
    DROP COLUMN IF EXISTS jwt_api_key_hash,
    DROP COLUMN IF EXISTS jwt_access_ttl_seconds,
    DROP COLUMN IF EXISTS jwt_refresh_ttl_seconds,
    DROP COLUMN IF EXISTS jwt_public_key_pem,
    DROP COLUMN IF EXISTS jwt_previous_public_key_pem,
    DROP COLUMN IF EXISTS jwt_key_id,
    DROP COLUMN IF EXISTS jwt_previous_key_id,
    DROP COLUMN IF EXISTS jwt_rotated_at,
    DROP COLUMN IF EXISTS jwt_rotate_every_days,
    DROP COLUMN IF EXISTS jwt_binding_mode,
    DROP COLUMN IF EXISTS jwt_max_sessions;
-- +goose StatementEnd
