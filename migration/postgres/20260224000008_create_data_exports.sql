-- +goose Up
CREATE TABLE IF NOT EXISTS data_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    file_url TEXT NOT NULL DEFAULT '',
    filters JSONB NOT NULL DEFAULT '{}',
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- +goose Down
DROP TABLE IF EXISTS data_exports;
