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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS session;
-- +goose StatementEnd
