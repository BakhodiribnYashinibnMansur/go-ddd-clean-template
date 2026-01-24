-- +goose Up
-- +goose StatementBegin

-- Add new columns to session table for detailed device/browser information
ALTER TABLE session 
ADD COLUMN IF NOT EXISTS os VARCHAR(100),
ADD COLUMN IF NOT EXISTS os_version VARCHAR(50),
ADD COLUMN IF NOT EXISTS browser VARCHAR(100),
ADD COLUMN IF NOT EXISTS browser_version VARCHAR(50);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_session_os ON session(os);
CREATE INDEX IF NOT EXISTS idx_session_browser ON session(browser);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove the added columns
ALTER TABLE session 
DROP COLUMN IF EXISTS browser_version,
DROP COLUMN IF EXISTS browser,
DROP COLUMN IF EXISTS os_version,
DROP COLUMN IF EXISTS os;

-- Drop the indexes
DROP INDEX IF EXISTS idx_session_browser;
DROP INDEX IF EXISTS idx_session_os;

-- +goose StatementEnd
