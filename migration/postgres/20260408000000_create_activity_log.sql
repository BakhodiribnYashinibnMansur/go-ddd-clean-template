-- +goose Up
-- +goose StatementBegin

-- =========================
-- ACTIVITY LOG TABLE
-- =========================
-- Field-level change tracking for business activity.
-- Each changed field in an update produces a separate row.
-- This is distinct from audit_log which tracks security/compliance events.
CREATE TABLE IF NOT EXISTS activity_log (
    id          BIGSERIAL PRIMARY KEY,
    actor_id    UUID NOT NULL,
    action      VARCHAR(64) NOT NULL,
    entity_type VARCHAR(64) NOT NULL,
    entity_id   UUID NOT NULL,
    field_name  VARCHAR(128),
    old_value   TEXT,
    new_value   TEXT,
    metadata    TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =========================
-- INDEXES FOR PERFORMANCE
-- =========================
CREATE INDEX IF NOT EXISTS idx_activity_log_entity  ON activity_log (entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_activity_log_actor   ON activity_log (actor_id);
CREATE INDEX IF NOT EXISTS idx_activity_log_action  ON activity_log (action);
CREATE INDEX IF NOT EXISTS idx_activity_log_created ON activity_log (created_at DESC);

-- Partial index for field-level queries
CREATE INDEX IF NOT EXISTS idx_activity_log_field ON activity_log (entity_type, field_name)
    WHERE field_name IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS activity_log;
-- +goose StatementEnd
