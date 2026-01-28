-- +goose Up
-- +goose StatementBegin

-- =========================
-- AUDIT ACTION TYPES ENUM
-- =========================
CREATE TYPE audit_action_type AS ENUM (
    'LOGIN',
    'LOGOUT',
    'SESSION_REVOKE',
    'PASSWORD_CHANGE',
    'MFA_VERIFY_FAIL',
    'ACCESS_GRANTED',
    'ACCESS_DENIED',
    'POLICY_MATCHED',
    'POLICY_DENIED',
    'USER_CREATE',
    'USER_UPDATE',
    'USER_DELETE',
    'ROLE_ASSIGN',
    'ROLE_REMOVE',
    'ORDER_APPROVE',
    'ORDER_CANCEL',
    'PAYMENT_PROCESS',
    'PAYMENT_CANCEL',
    'POLICY_EVALUATED'
);

-- =========================
-- AUDIT LOG TABLE
-- =========================
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- who
    user_id UUID REFERENCES users(id),
    session_id UUID REFERENCES session(id),

    -- what
    action audit_action_type NOT NULL,
    resource_type VARCHAR(64),          -- user, order, role, policy, etc.
    resource_id UUID,

    -- context
    platform VARCHAR(16),               -- admin / web / mobile / api
    ip_address INET,
    user_agent VARCHAR(512),

    -- authz info
    permission VARCHAR(128),            -- permission being evaluated
    policy_id UUID REFERENCES policy(id), -- policy that made the decision
    decision VARCHAR(16),               -- ALLOW / DENY

    -- result
    success BOOLEAN NOT NULL,
    error_message TEXT,

    -- extra metadata for debugging and compliance
    metadata JSONB,                     -- request_id, diff, payload snapshot, etc.

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- =========================
-- INDEXES FOR PERFORMANCE
-- =========================
CREATE INDEX idx_audit_user_id ON audit_log(user_id);
CREATE INDEX idx_audit_session_id ON audit_log(session_id);
CREATE INDEX idx_audit_action ON audit_log(action);
CREATE INDEX idx_audit_resource ON audit_log(resource_type, resource_id);
CREATE INDEX idx_audit_created_at ON audit_log(created_at);
CREATE INDEX idx_audit_decision ON audit_log(decision) WHERE decision IS NOT NULL;
CREATE INDEX idx_audit_policy_id ON audit_log(policy_id) WHERE policy_id IS NOT NULL;

-- Partial index for failed attempts (security monitoring)
CREATE INDEX idx_audit_failed_attempts ON audit_log(created_at, user_id, action) 
WHERE success = FALSE;

-- =========================
-- SECURITY: RESTRICT ACCESS
-- =========================
-- Only allow auditor_role to read audit logs (implement role-based access at application level)
-- This table should be append-only - no updates or deletes allowed through normal operations

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS audit_log;
DROP TYPE IF EXISTS audit_action_type;
-- +goose StatementEnd
