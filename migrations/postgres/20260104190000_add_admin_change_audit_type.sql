-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'ADMIN_CHANGE' AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'audit_action_type')) THEN
        ALTER TYPE audit_action_type ADD VALUE 'ADMIN_CHANGE';
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Removing a value from an ENUM is not directly supported in Postgres.
-- This would require creating a new type, migrating data, and dropping the old type.
-- Since this is an additive change, we usually leave it.
-- +goose StatementEnd
