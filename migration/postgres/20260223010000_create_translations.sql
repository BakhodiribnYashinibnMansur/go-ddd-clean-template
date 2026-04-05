-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS translations (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL,
    entity_id   UUID        NOT NULL,
    lang_code   VARCHAR(5)  NOT NULL,   -- 'uz', 'ru', 'en'
    data        JSONB       NOT NULL DEFAULT '{}',  -- {"title": "...", "description": "..."}
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (entity_type, entity_id, lang_code)
);

CREATE INDEX IF NOT EXISTS idx_translations_lookup
    ON translations (entity_type, entity_id, lang_code);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_translations_lookup;
DROP TABLE IF EXISTS translations;
-- +goose StatementEnd
