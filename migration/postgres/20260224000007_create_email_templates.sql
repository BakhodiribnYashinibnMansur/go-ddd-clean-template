-- +goose Up
CREATE TABLE IF NOT EXISTS email_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    subject TEXT NOT NULL,
    html_body TEXT NOT NULL DEFAULT '',
    text_body TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL DEFAULT 'transactional',
    is_active BOOL NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS email_templates;
