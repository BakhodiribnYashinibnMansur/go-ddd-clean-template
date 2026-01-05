-- +goose Up
-- +goose StatementBegin
ALTER TYPE audit_action_type ADD VALUE 'ADMIN_CHANGE';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Removing a value from an ENUM is not directly supported in Postgres.
-- This would require creating a new type, migrating data, and dropping the old type.
-- Since this is an additive change, we usually leave it.
-- +goose StatementEnd
