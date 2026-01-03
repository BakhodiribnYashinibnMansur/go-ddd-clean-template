-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS session (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    
    device_id TEXT NOT NULL,
    device_name TEXT,
    device_type TEXT CHECK(device_type IN ('DESKTOP', 'MOBILE', 'TABLET', 'BOT', 'TV')),
    
    ip_address TEXT,
    user_agent TEXT,
    
    fcm_token TEXT,
    refresh_token_hash TEXT,
    
    expires_at DATETIME NOT NULL,
    last_activity DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked INTEGER NOT NULL DEFAULT 0,
    
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE
);

CREATE INDEX idx_session_user_id ON session(user_id);
CREATE INDEX idx_session_device_id ON session(device_id);
CREATE INDEX idx_session_expires_at ON session(expires_at);
CREATE INDEX idx_session_last_activity ON session(last_activity);
CREATE INDEX idx_session_revoked ON session(revoked) WHERE revoked = 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_session_revoked;
DROP INDEX IF EXISTS idx_session_last_activity;
DROP INDEX IF EXISTS idx_session_expires_at;
DROP INDEX IF EXISTS idx_session_device_id;
DROP INDEX IF EXISTS idx_session_user_id;
DROP TABLE IF EXISTS session;
-- +goose StatementEnd
