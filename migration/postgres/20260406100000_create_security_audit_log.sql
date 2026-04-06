-- +goose Up
CREATE TABLE IF NOT EXISTS security_audit_log (
    id BIGSERIAL PRIMARY KEY,
    event TEXT NOT NULL,
    integration_name TEXT,
    user_id UUID,
    session_id UUID,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_security_audit_log_event ON security_audit_log(event);
CREATE INDEX idx_security_audit_log_user_id ON security_audit_log(user_id);
CREATE INDEX idx_security_audit_log_created_at ON security_audit_log(created_at);
CREATE INDEX idx_security_audit_log_integration ON security_audit_log(integration_name);

-- Prevent modifications: only INSERT is intended.
-- (Application-level enforcement — no REVOKE since we don't control the DB role here.)

-- +goose Down
DROP TABLE IF EXISTS security_audit_log;
