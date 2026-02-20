-- +goose Up
-- +goose StatementBegin

-- Add cache invalidation triggers for integrations table
CREATE TRIGGER invalidate_cache_integrations
AFTER INSERT OR UPDATE OR DELETE ON integrations
FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

-- Add cache invalidation triggers for api_keys table
CREATE TRIGGER invalidate_cache_api_keys
AFTER INSERT OR UPDATE OR DELETE ON api_keys
FOR EACH ROW EXECUTE FUNCTION notify_cache_invalidation();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS invalidate_cache_api_keys ON api_keys;
DROP TRIGGER IF EXISTS invalidate_cache_integrations ON integrations;
-- +goose StatementEnd
