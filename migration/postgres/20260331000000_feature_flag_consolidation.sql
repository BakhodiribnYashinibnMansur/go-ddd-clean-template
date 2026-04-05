-- +goose Up

-- Add new columns to feature_flags
ALTER TABLE feature_flags ADD COLUMN IF NOT EXISTS rollout_percentage INT NOT NULL DEFAULT 0;
ALTER TABLE feature_flags ADD COLUMN IF NOT EXISTS default_value TEXT NOT NULL DEFAULT '';

-- Rename 'type' to 'flag_type' for Go keyword safety
ALTER TABLE feature_flags RENAME COLUMN type TO flag_type;

-- Create rule groups table
CREATE TABLE IF NOT EXISTS feature_flag_rule_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    flag_id UUID NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    variation TEXT NOT NULL,
    priority INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ff_rule_groups_flag_priority ON feature_flag_rule_groups(flag_id, priority);

-- Create conditions table
CREATE TABLE IF NOT EXISTS feature_flag_conditions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_group_id UUID NOT NULL REFERENCES feature_flag_rule_groups(id) ON DELETE CASCADE,
    attribute TEXT NOT NULL,
    operator TEXT NOT NULL CHECK (operator IN ('eq','not_eq','in','not_in','gt','gte','lt','lte','contains')),
    value TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_ff_conditions_rule_group ON feature_flag_conditions(rule_group_id);

-- +goose Down
DROP TABLE IF EXISTS feature_flag_conditions;
DROP TABLE IF EXISTS feature_flag_rule_groups;
ALTER TABLE feature_flags RENAME COLUMN flag_type TO type;
ALTER TABLE feature_flags DROP COLUMN IF EXISTS rollout_percentage;
ALTER TABLE feature_flags DROP COLUMN IF EXISTS default_value;
