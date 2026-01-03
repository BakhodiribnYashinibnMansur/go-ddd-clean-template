-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

DROP TABLE IF EXISTS session CASCADE;
DROP TABLE IF EXISTS user_relation CASCADE;
DROP TABLE IF EXISTS "user" CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- =========================
-- ENUM TYPES
-- =========================
CREATE TYPE relation_types AS ENUM ('UNREVEALED', 'BRANCH', 'REGION');
CREATE TYPE policy_effect AS ENUM ('ALLOW', 'DENY');

-- =========================
-- ROLE / PERMISSION (RBAC)
-- =========================
CREATE TABLE role (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (name)
);

CREATE TABLE permission (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID REFERENCES permission(id),
    name VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (parent_id, name)
);

CREATE TABLE role_permission (
    role_id UUID REFERENCES role(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permission(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, permission_id)
);

-- =========================
-- API SCOPE (endpoint-level)
-- =========================
CREATE TABLE scope (
    path VARCHAR NOT NULL,
    method VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (path, method)
);

CREATE TABLE permission_scope (
    permission_id UUID REFERENCES permission(id) ON DELETE CASCADE,
    path VARCHAR,
    method VARCHAR,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (permission_id, path, method),
    FOREIGN KEY (path, method)
        REFERENCES scope(path, method)
);

-- =========================
-- USER (RBAC + ABAC CONTEXT)
-- =========================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID REFERENCES role(id),
    username VARCHAR,
    email VARCHAR,
    phone VARCHAR,
    password_hash TEXT,
    salt VARCHAR,                     -- merged from old users table
    attributes JSONB NOT NULL DEFAULT '{}', -- region, branch, dept
    active BOOLEAN DEFAULT TRUE,
    last_seen TIMESTAMP,              -- merged from old users table
    deleted_at BIGINT DEFAULT 0,      -- merged from old users table
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (email),
    UNIQUE (username),
    UNIQUE (phone)
);

CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- =========================
-- SESSION
-- =========================
CREATE TYPE session_device_type AS ENUM ('DESKTOP', 'MOBILE', 'TABLET', 'BOT', 'TV');

CREATE TABLE session (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id UUID NOT NULL,
    device_name VARCHAR(255),
    device_type session_device_type,
    ip_address INET,
    user_agent VARCHAR(512),
    fcm_token VARCHAR(512),
    data JSONB,
    refresh_token_hash VARCHAR(512),
    expires_at TIMESTAMP NOT NULL,
    last_activity TIMESTAMP NOT NULL DEFAULT NOW(),
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_session_user_id ON session(user_id);
CREATE INDEX idx_session_device_id ON session(device_id);
CREATE INDEX idx_session_expires_at ON session(expires_at);
CREATE INDEX idx_session_last_activity ON session(last_activity);
CREATE INDEX idx_session_revoked ON session(revoked) WHERE revoked = FALSE;

-- =========================
-- RELATION (ORG STRUCTURE)
-- =========================
CREATE TABLE relation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type relation_types NOT NULL, -- REGION / BRANCH
    name VARCHAR NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE (type, name)
);

CREATE TABLE user_relation (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    relation_id UUID REFERENCES relation(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, relation_id)
);


-- =========================
-- POLICY (ABAC - JSON CONDITIONS)
-- =========================
CREATE TABLE policy (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    permission_id UUID REFERENCES permission(id) ON DELETE CASCADE,
    effect policy_effect NOT NULL,      -- ALLOW / DENY
    priority INT DEFAULT 100,
    active BOOLEAN DEFAULT TRUE,
    conditions JSONB NOT NULL,          -- ABAC rules
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_policy_permission_id ON policy(permission_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS policy;
DROP TABLE IF EXISTS session;
DROP TABLE IF EXISTS user_relation;
DROP TABLE IF EXISTS relation;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS permission_scope;
DROP TABLE IF EXISTS scope;
DROP TABLE IF EXISTS role_permission;
DROP TABLE IF EXISTS permission;
DROP TABLE IF EXISTS role;

DROP TYPE IF EXISTS session_device_type;
DROP TYPE IF EXISTS policy_effect;
DROP TYPE IF EXISTS relation_types;
-- +goose StatementEnd
