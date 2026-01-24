-- +goose Up
-- +goose StatementBegin
CREATE TABLE site_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(64) UNIQUE NOT NULL,
    value TEXT,
    value_type VARCHAR(16) NOT NULL DEFAULT 'string', -- string, boolean, integer, json
    category VARCHAR(32) NOT NULL DEFAULT 'general', -- general, email, maintenance, api
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE, -- Can be accessed without auth
    
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index for fast lookups
CREATE INDEX idx_site_settings_key ON site_settings(key);
CREATE INDEX idx_site_settings_category ON site_settings(category);
CREATE INDEX idx_site_settings_public ON site_settings(is_public);

-- Insert default settings
INSERT INTO site_settings (key, value, value_type, category, description, is_public) VALUES
    ('site_name', 'Go Clean Template', 'string', 'general', 'Application name', true),
    ('site_description', 'A clean architecture template for Go applications', 'string', 'general', 'Site description', true),
    ('maintenance_mode', 'false', 'boolean', 'maintenance', 'Enable maintenance mode', false),
    ('maintenance_message', 'We are currently performing maintenance. Please check back soon.', 'string', 'maintenance', 'Maintenance mode message', true),
    ('allow_registration', 'true', 'boolean', 'general', 'Allow new user registration', false),
    ('max_upload_size', '10485760', 'integer', 'general', 'Maximum upload size in bytes (10MB)', false),
    ('session_timeout', '3600', 'integer', 'general', 'Session timeout in seconds', false),
    ('admin_email', 'admin@example.com', 'string', 'email', 'Admin email address', false),
    ('smtp_enabled', 'false', 'boolean', 'email', 'Enable SMTP email sending', false),
    ('items_per_page', '10', 'integer', 'general', 'Default items per page', false);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS site_settings;
-- +goose StatementEnd
