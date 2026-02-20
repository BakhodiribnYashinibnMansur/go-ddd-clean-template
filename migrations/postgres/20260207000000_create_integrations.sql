-- +goose Up
-- +goose StatementBegin

-- Create integrations table
CREATE TABLE IF NOT EXISTS integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    base_url VARCHAR(500) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    config JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create API keys table
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_id UUID NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    key VARCHAR(255) NOT NULL UNIQUE, -- Hashed API key
    key_prefix VARCHAR(20) NOT NULL, -- Visible prefix (e.g., "sk_live_...")
    is_active BOOLEAN NOT NULL DEFAULT true,
    expires_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_integrations_name ON integrations(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_integrations_is_active ON integrations(is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_integration_id ON api_keys(integration_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_api_keys_key ON api_keys(key) WHERE deleted_at IS NULL AND is_active = true;
CREATE INDEX IF NOT EXISTS idx_api_keys_key_prefix ON api_keys(key_prefix) WHERE deleted_at IS NULL;

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_integrations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_api_keys_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_integrations_updated_at
    BEFORE UPDATE ON integrations
    FOR EACH ROW
    EXECUTE FUNCTION update_integrations_updated_at();

CREATE TRIGGER trigger_api_keys_updated_at
    BEFORE UPDATE ON api_keys
    FOR EACH ROW
    EXECUTE FUNCTION update_api_keys_updated_at();

-- Add comments for documentation
COMMENT ON TABLE integrations IS 'Stores third-party integration platform configurations';
COMMENT ON TABLE api_keys IS 'Stores API keys for accessing integrations';
COMMENT ON COLUMN api_keys.key IS 'Hashed API key for security';
COMMENT ON COLUMN api_keys.key_prefix IS 'Visible prefix for key identification (e.g., sk_live_abc...)';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS api_keys CASCADE;
DROP TABLE IF EXISTS integrations CASCADE;
DROP FUNCTION IF EXISTS update_api_keys_updated_at() CASCADE;
DROP FUNCTION IF EXISTS update_integrations_updated_at() CASCADE;
-- +goose StatementEnd
