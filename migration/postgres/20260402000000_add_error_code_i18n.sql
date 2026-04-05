-- +goose Up
-- +goose StatementBegin
ALTER TABLE error_code
    ADD COLUMN IF NOT EXISTS message_uz TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS message_ru TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE error_code
    DROP COLUMN IF EXISTS message_uz,
    DROP COLUMN IF EXISTS message_ru;
-- +goose StatementEnd
