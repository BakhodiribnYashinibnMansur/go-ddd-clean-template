-- +goose Up
CREATE TABLE IF NOT EXISTS file_metadata (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    original_name TEXT NOT NULL,
    stored_name TEXT NOT NULL,
    bucket TEXT NOT NULL,
    url TEXT NOT NULL,
    size BIGINT NOT NULL DEFAULT 0,
    mime_type TEXT NOT NULL DEFAULT '',
    uploaded_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS file_metadata;
