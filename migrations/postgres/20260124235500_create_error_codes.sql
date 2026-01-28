-- +goose Up
-- +goose StatementBegin
CREATE TYPE error_severity_enum AS ENUM ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL');
CREATE TYPE error_category_enum AS ENUM ('DATA', 'AUTH', 'SYSTEM', 'VALIDATION', 'BUSINESS', 'UNKNOWN');

CREATE TABLE error_code (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(255) NOT NULL UNIQUE,
    message TEXT NOT NULL,
    http_status INT NOT NULL,
    
    -- Additional metadata
    category error_category_enum DEFAULT 'UNKNOWN',
    severity error_severity_enum DEFAULT 'MEDIUM',
    retryable BOOLEAN DEFAULT FALSE,
    retry_after INT DEFAULT 0,
    suggestion TEXT,
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_error_code_code ON error_code(code);
CREATE INDEX idx_error_code_category ON error_code(category);

-- Seed initial error code
INSERT INTO error_code (code, message, http_status, category, severity, retryable, retry_after, suggestion)
VALUES (
    'RESOURCE_NOT_FOUND',
    'The requested resource was not found.',
    404,
    'DATA',
    'MEDIUM',
    FALSE,
    5,
    'Please check our documentation.'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS error_code;
DROP TYPE IF EXISTS error_category_enum;
DROP TYPE IF EXISTS error_severity_enum;
-- +goose StatementEnd
