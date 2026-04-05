-- +goose Up
CREATE TABLE IF NOT EXISTS ip_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ip_address TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('allow', 'block')),
    reason TEXT NOT NULL DEFAULT '',
    is_active BOOL NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS ip_rules;
