-- +goose Up
-- +goose StatementBegin
-- JSONB columns removed: conditions and attributes now use entity_metadata (EAV).
-- This migration is intentionally empty after the JSONB-to-EAV migration.
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- No-op
-- +goose StatementEnd
