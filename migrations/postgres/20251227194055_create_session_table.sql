-- +goose Up
-- +goose StatementBegin
CREATE TYPE session_device_type AS ENUM
('DESKTOP', 'MOBILE', 'TABLET', 'BOT', 'TV');

CREATE TABLE IF NOT EXISTS session (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    turon_id BIGINT NOT NULL,

    device_id UUID NOT NULL,
    device_name VARCHAR(255),
    device_type session_device_type,

    ip_address INET,
    user_agent VARCHAR(512),

    fcm_token VARCHAR(512),

    expires_at TIMESTAMP NOT NULL,
    last_activity TIMESTAMP NOT NULL DEFAULT NOW(),

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    FOREIGN KEY (turon_id)
        REFERENCES "user"(turon_id)
        ON DELETE CASCADE
);

CREATE INDEX idx_session_turon_id ON session(turon_id);
CREATE INDEX idx_session_device_id ON session(device_id);
CREATE INDEX idx_session_expires_at ON session(expires_at);
CREATE INDEX idx_session_last_activity ON session(last_activity);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS session;
DROP TYPE IF EXISTS session_device_type;
-- +goose StatementEnd
