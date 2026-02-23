-- +goose Up
-- +goose StatementBegin
ALTER TABLE role ADD COLUMN IF NOT EXISTS description VARCHAR;
ALTER TABLE permission ADD COLUMN IF NOT EXISTS description VARCHAR;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE role DROP COLUMN IF EXISTS description;
ALTER TABLE permission DROP COLUMN IF EXISTS description;
-- +goose StatementEnd
