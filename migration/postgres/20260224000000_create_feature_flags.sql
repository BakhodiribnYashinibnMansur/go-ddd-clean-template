-- +goose Up
CREATE TABLE IF NOT EXISTS feature_flags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('bool', 'string', 'int', 'json')),
    value TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    is_active BOOL NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_feature_flags_key ON feature_flags(key);

-- +goose Down
DROP TABLE IF EXISTS feature_flags;
