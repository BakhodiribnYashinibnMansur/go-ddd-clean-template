-- +goose Up
CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    cron_schedule TEXT NOT NULL DEFAULT '',
    payload JSONB NOT NULL DEFAULT '{}',
    is_active BOOL NOT NULL DEFAULT true,
    status TEXT NOT NULL DEFAULT 'idle',
    last_run_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS jobs;
