-- +goose Up
-- +goose StatementBegin

-- Create the universal entity_metadata EAV table early so that seed migrations can reference it.
CREATE TABLE IF NOT EXISTS entity_metadata (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(64)  NOT NULL,
    entity_id   UUID         NOT NULL,
    key         VARCHAR(128) NOT NULL,
    value       TEXT         NOT NULL DEFAULT '',
    value_type  VARCHAR(16)  NOT NULL DEFAULT 'string',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (entity_type, entity_id, key)
);
CREATE INDEX IF NOT EXISTS idx_entity_metadata_lookup ON entity_metadata(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_entity_metadata_type ON entity_metadata(entity_type);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Intentionally empty: the table is dropped in 20260401000000_jsonb_to_eav.sql down migration.

-- +goose StatementEnd
