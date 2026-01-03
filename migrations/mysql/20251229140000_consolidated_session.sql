-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS session (
    id CHAR(36) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    
    device_id CHAR(36) NOT NULL,
    device_name VARCHAR(255),
    device_type ENUM('DESKTOP', 'MOBILE', 'TABLET', 'BOT', 'TV'),
    
    ip_address VARCHAR(45),
    user_agent VARCHAR(512),
    
    fcm_token VARCHAR(512),
    refresh_token_hash VARCHAR(255),
    
    expires_at TIMESTAMP NOT NULL,
    last_activity TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_session_user_id (user_id),
    INDEX idx_session_device_id (device_id),
    INDEX idx_session_expires_at (expires_at),
    INDEX idx_session_last_activity (last_activity),
    INDEX idx_session_revoked (revoked),
    
    FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS session;
-- +goose StatementEnd
