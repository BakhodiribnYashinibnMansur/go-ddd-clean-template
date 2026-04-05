-- +goose Up
-- +goose StatementBegin

-- outbox_events implements the transactional outbox pattern. A producer
-- writes the row inside the same transaction as the aggregate change; a
-- relay goroutine publishes rows to the event bus and marks dispatched_at,
-- giving at-least-once delivery across bus/crash failures.
CREATE TABLE IF NOT EXISTS outbox_events (
    id             BIGSERIAL   PRIMARY KEY,
    event_id       UUID        NOT NULL UNIQUE,
    event_name     TEXT        NOT NULL,
    aggregate_id   UUID        NOT NULL,
    payload        BYTEA       NOT NULL,
    occurred_at    TIMESTAMPTZ NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    dispatched_at  TIMESTAMPTZ NULL,
    attempts       INTEGER     NOT NULL DEFAULT 0,
    last_error     TEXT        NULL
);

-- Relay scans undispatched rows ordered by id.
CREATE INDEX IF NOT EXISTS idx_outbox_events_undispatched
    ON outbox_events (id) WHERE dispatched_at IS NULL;

-- Observability: quickly find rows by event name or aggregate.
CREATE INDEX IF NOT EXISTS idx_outbox_events_event_name  ON outbox_events (event_name);
CREATE INDEX IF NOT EXISTS idx_outbox_events_aggregate   ON outbox_events (aggregate_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS outbox_events;
-- +goose StatementEnd
