-- +goose Up
-- +goose StatementBegin

-- 1. Create Trigger Function used to notify cache invalidation listeners
CREATE OR REPLACE FUNCTION notify_cache_invalidation() RETURNS TRIGGER AS $$
DECLARE
    payload TEXT;
BEGIN
    -- Payload is the table name
    payload := TG_TABLE_NAME;
    -- Send notification to channel 'cache_invalidation'
    PERFORM pg_notify('cache_invalidation', payload);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- 2. Create Triggers for key tables
-- Users table
CREATE TRIGGER invalidate_cache_users
AFTER INSERT OR UPDATE OR DELETE ON users
FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

-- Role table
CREATE TRIGGER invalidate_cache_role
AFTER INSERT OR UPDATE OR DELETE ON role
FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

-- Permission table
CREATE TRIGGER invalidate_cache_permission
AFTER INSERT OR UPDATE OR DELETE ON permission
FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

-- Policy table
CREATE TRIGGER invalidate_cache_policy
AFTER INSERT OR UPDATE OR DELETE ON policy
FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

-- Session table (optional, but good for security)
CREATE TRIGGER invalidate_cache_session
AFTER INSERT OR UPDATE OR DELETE ON session
FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

-- Relation table
CREATE TRIGGER invalidate_cache_relation
AFTER INSERT OR UPDATE OR DELETE ON relation
FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS invalidate_cache_relation ON relation;
DROP TRIGGER IF EXISTS invalidate_cache_session ON session;
DROP TRIGGER IF EXISTS invalidate_cache_policy ON policy;
DROP TRIGGER IF EXISTS invalidate_cache_permission ON permission;
DROP TRIGGER IF EXISTS invalidate_cache_role ON role;
DROP TRIGGER IF EXISTS invalidate_cache_users ON users;

DROP FUNCTION IF EXISTS notify_cache_invalidation;
-- +goose StatementEnd
