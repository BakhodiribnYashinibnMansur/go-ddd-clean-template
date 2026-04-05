-- +goose Up
CREATE TABLE IF NOT EXISTS rate_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    path_pattern TEXT NOT NULL,
    method TEXT NOT NULL DEFAULT 'ALL',
    limit_count INT NOT NULL,
    window_seconds INT NOT NULL,
    is_active BOOL NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS rate_limits;
