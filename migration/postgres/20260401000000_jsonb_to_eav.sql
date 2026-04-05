-- +goose Up
-- +goose StatementBegin

-- 1. Create the universal entity_metadata EAV table.
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

-- 2. Migrate data from JSONB columns into entity_metadata.

-- users.attributes
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'user_attributes', id, kv.key, kv.value, 'string'
FROM users, jsonb_each_text(attributes) AS kv
WHERE attributes IS NOT NULL AND attributes != '{}'::jsonb
ON CONFLICT DO NOTHING;

-- policy.conditions (may contain arrays, so use jsonb_each not jsonb_each_text)
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'policy_conditions', id, kv.key,
    CASE
        WHEN jsonb_typeof(kv.value) = 'array' THEN kv.value::text
        ELSE kv.value #>> '{}'
    END,
    CASE
        WHEN jsonb_typeof(kv.value) = 'array' THEN 'json_array'
        ELSE 'string'
    END
FROM policy, jsonb_each(conditions) AS kv
WHERE conditions IS NOT NULL AND conditions != '{}'::jsonb
ON CONFLICT DO NOTHING;

-- audit_log.metadata
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'audit_log_metadata', id, kv.key, kv.value, 'string'
FROM audit_log, jsonb_each_text(metadata) AS kv
WHERE metadata IS NOT NULL AND metadata != '{}'::jsonb
ON CONFLICT DO NOTHING;

-- system_errors.metadata
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'system_error_metadata', id, kv.key, kv.value, 'string'
FROM system_errors, jsonb_each_text(metadata) AS kv
WHERE metadata IS NOT NULL AND metadata != '{}'::jsonb
ON CONFLICT DO NOTHING;

-- integrations.config
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'integration_config', id, kv.key, kv.value, 'string'
FROM integrations, jsonb_each_text(config) AS kv
WHERE config IS NOT NULL AND config != '{}'::jsonb
ON CONFLICT DO NOTHING;

-- jobs.payload
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'job_payload', id, kv.key, kv.value, 'string'
FROM jobs, jsonb_each_text(payload) AS kv
WHERE payload IS NOT NULL AND payload != '{}'::jsonb
ON CONFLICT DO NOTHING;

-- translations.data
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'translation_data', id, kv.key, kv.value, 'string'
FROM translations, jsonb_each_text(data) AS kv
WHERE data IS NOT NULL AND data != '{}'::jsonb
ON CONFLICT DO NOTHING;

-- data_exports.filters
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'data_export_filters', id, kv.key, kv.value, 'string'
FROM data_exports, jsonb_each_text(filters) AS kv
WHERE filters IS NOT NULL AND filters != '{}'::jsonb
ON CONFLICT DO NOTHING;

-- session.data
INSERT INTO entity_metadata (entity_type, entity_id, key, value, value_type)
SELECT 'session_data', id, kv.key, kv.value, 'string'
FROM session, jsonb_each_text(data) AS kv
WHERE data IS NOT NULL AND data != '{}'::jsonb
ON CONFLICT DO NOTHING;

-- 3. Drop JSONB columns.
ALTER TABLE users DROP COLUMN IF EXISTS attributes;
ALTER TABLE session DROP COLUMN IF EXISTS data;
ALTER TABLE policy DROP COLUMN IF EXISTS conditions;
ALTER TABLE audit_log DROP COLUMN IF EXISTS metadata;
ALTER TABLE system_errors DROP COLUMN IF EXISTS metadata;
ALTER TABLE integrations DROP COLUMN IF EXISTS config;
ALTER TABLE jobs DROP COLUMN IF EXISTS payload;
ALTER TABLE translations DROP COLUMN IF EXISTS data;
ALTER TABLE data_exports DROP COLUMN IF EXISTS filters;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Re-add JSONB columns.
ALTER TABLE users ADD COLUMN IF NOT EXISTS attributes JSONB NOT NULL DEFAULT '{}';
ALTER TABLE session ADD COLUMN IF NOT EXISTS data JSONB;
ALTER TABLE policy ADD COLUMN IF NOT EXISTS conditions JSONB NOT NULL DEFAULT '{}';
ALTER TABLE audit_log ADD COLUMN IF NOT EXISTS metadata JSONB;
ALTER TABLE system_errors ADD COLUMN IF NOT EXISTS metadata JSONB;
ALTER TABLE integrations ADD COLUMN IF NOT EXISTS config JSONB DEFAULT '{}'::jsonb;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS payload JSONB NOT NULL DEFAULT '{}';
ALTER TABLE translations ADD COLUMN IF NOT EXISTS data JSONB NOT NULL DEFAULT '{}';
ALTER TABLE data_exports ADD COLUMN IF NOT EXISTS filters JSONB NOT NULL DEFAULT '{}';

-- Migrate data back from entity_metadata to JSONB columns.
UPDATE users u SET attributes = COALESCE((
    SELECT jsonb_object_agg(key, value) FROM entity_metadata
    WHERE entity_type = 'user_attributes' AND entity_id = u.id
), '{}');

UPDATE policy p SET conditions = COALESCE((
    SELECT jsonb_object_agg(key,
        CASE WHEN value_type = 'json_array' THEN value::jsonb ELSE to_jsonb(value) END
    ) FROM entity_metadata
    WHERE entity_type = 'policy_conditions' AND entity_id = p.id
), '{}');

UPDATE audit_log a SET metadata = COALESCE((
    SELECT jsonb_object_agg(key, value) FROM entity_metadata
    WHERE entity_type = 'audit_log_metadata' AND entity_id = a.id
), '{}');

UPDATE system_errors s SET metadata = COALESCE((
    SELECT jsonb_object_agg(key, value) FROM entity_metadata
    WHERE entity_type = 'system_error_metadata' AND entity_id = s.id
), '{}');

UPDATE integrations i SET config = COALESCE((
    SELECT jsonb_object_agg(key, value) FROM entity_metadata
    WHERE entity_type = 'integration_config' AND entity_id = i.id
), '{}');

DROP TABLE IF EXISTS entity_metadata;

-- +goose StatementEnd
