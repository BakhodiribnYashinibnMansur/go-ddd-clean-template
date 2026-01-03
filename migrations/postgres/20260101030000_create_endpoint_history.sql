-- +goose Up
-- +goose StatementBegin

-- =========================
-- ENDPOINT HISTORY TABLE
-- =========================
CREATE TABLE endpoint_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- identity
    user_id UUID REFERENCES users(id),
    session_id UUID REFERENCES session(id),

    -- request
    method VARCHAR(8) NOT NULL,
    path VARCHAR(255) NOT NULL,

    -- response
    status_code SMALLINT NOT NULL,
    duration_ms INTEGER NOT NULL,

    -- context
    platform VARCHAR(16),               -- admin / web / mobile / api
    ip_address INET,
    user_agent VARCHAR(512),

    -- authz context
    permission VARCHAR(128),            -- permission being evaluated
    decision VARCHAR(16),               -- ALLOW / DENY

    -- meta
    request_id UUID,
    rate_limited BOOLEAN DEFAULT FALSE,
    response_size INTEGER,              -- response body size in bytes
    error_message TEXT,                 -- only for 5xx errors

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- =========================
-- INDEXES FOR PERFORMANCE
-- =========================
CREATE INDEX idx_eh_user_id ON endpoint_history(user_id);
CREATE INDEX idx_eh_session_id ON endpoint_history(session_id);
CREATE INDEX idx_eh_path ON endpoint_history(path);
CREATE INDEX idx_eh_method ON endpoint_history(method);
CREATE INDEX idx_eh_status ON endpoint_history(status_code);
CREATE INDEX idx_eh_created_at ON endpoint_history(created_at);

-- Composite indexes for common queries
CREATE INDEX idx_eh_user_created ON endpoint_history(user_id, created_at);
CREATE INDEX idx_eh_path_status ON endpoint_history(path, status_code);
CREATE INDEX idx_eh_decision ON endpoint_history(decision) WHERE decision IS NOT NULL;

-- Partial index for errors (monitoring)
CREATE INDEX idx_eh_errors ON endpoint_history(created_at, path, status_code) 
WHERE status_code >= 500;

-- Partial index for slow requests (performance monitoring)
CREATE INDEX idx_eh_slow_requests ON endpoint_history(created_at, path, duration_ms) 
WHERE duration_ms > 1000;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS endpoint_history;
-- +goose StatementEnd
