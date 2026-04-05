-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS app_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    level VARCHAR(8) NOT NULL,         -- debug, info, warn, error, fatal
    message TEXT NOT NULL,
    caller VARCHAR(255),               -- file:line -> function

    -- Structured fields
    operation VARCHAR(128),
    entity VARCHAR(64),
    entity_id VARCHAR(128),
    error_text TEXT,

    -- Context
    request_id VARCHAR(64),
    user_id VARCHAR(128),
    session_id VARCHAR(64),
    ip_address VARCHAR(45),

    -- Raw extra fields as text (not JSONB per project convention)
    extra TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_app_logs_level ON app_logs(level);
CREATE INDEX IF NOT EXISTS idx_app_logs_created_at ON app_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_app_logs_request_id ON app_logs(request_id) WHERE request_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_app_logs_operation ON app_logs(operation) WHERE operation IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_app_logs_entity ON app_logs(entity) WHERE entity IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS app_logs;
-- +goose StatementEnd
