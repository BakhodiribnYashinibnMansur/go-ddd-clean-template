-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS system_errors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(64) NOT NULL,
    message TEXT NOT NULL,
    stack_trace TEXT,
    metadata JSONB,
    severity VARCHAR(16) NOT NULL DEFAULT 'ERROR', -- ERROR, FATAL, PANIC, WARN
    service_name VARCHAR(64) DEFAULT 'api',
    
    -- Context
    request_id UUID,
    user_id UUID,
    ip_address INET,
    path VARCHAR(255),
    method VARCHAR(8),
    
    is_resolved BOOLEAN DEFAULT FALSE,
    resolved_at TIMESTAMP,
    resolved_by UUID,
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sys_err_code ON system_errors(code);
CREATE INDEX IF NOT EXISTS idx_sys_err_severity ON system_errors(severity);
CREATE INDEX IF NOT EXISTS idx_sys_err_created_at ON system_errors(created_at);
CREATE INDEX IF NOT EXISTS idx_sys_err_req_id ON system_errors(request_id) WHERE request_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_sys_err_resolved ON system_errors(is_resolved);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS system_errors;
-- +goose StatementEnd
