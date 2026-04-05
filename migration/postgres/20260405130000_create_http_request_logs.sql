-- +goose Up
-- +goose StatementBegin

-- http_request_logs is a high-write, time-series table. It is partitioned by
-- month on created_at so retention is an O(1) DROP TABLE rather than a full
-- DELETE that rewrites the heap. Indexes are partial to keep them small on
-- skewed columns (response_status: 95% of rows are 200).
CREATE TABLE IF NOT EXISTS http_request_logs (
    id UUID NOT NULL DEFAULT gen_random_uuid(),

    method              VARCHAR(10)  NOT NULL,
    path                TEXT         NOT NULL,
    query               TEXT,
    route               VARCHAR(255),              -- matched Gin route pattern

    request_headers     TEXT,
    request_body        TEXT,
    request_body_size   INT,

    response_status     INT          NOT NULL,
    response_headers    TEXT,
    response_body       TEXT,
    response_body_size  INT,

    duration_ms         INT          NOT NULL,
    client_ip           VARCHAR(45),
    user_agent          TEXT,

    request_id          VARCHAR(64),
    user_id             VARCHAR(128),
    session_id          VARCHAR(64),

    created_at          TIMESTAMP    NOT NULL DEFAULT NOW(),

    -- Partition key MUST be part of the primary key.
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Error-path partial index — 5% of traffic, queried most often for debugging.
CREATE INDEX IF NOT EXISTS idx_http_req_logs_errors
    ON http_request_logs (created_at DESC, response_status)
    WHERE response_status >= 400;

-- Slow-path partial index — operators hunting latency regressions.
CREATE INDEX IF NOT EXISTS idx_http_req_logs_slow
    ON http_request_logs (created_at DESC, duration_ms)
    WHERE duration_ms >= 500;

-- Correlation lookups.
CREATE INDEX IF NOT EXISTS idx_http_req_logs_request_id
    ON http_request_logs (request_id)
    WHERE request_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_http_req_logs_user_id
    ON http_request_logs (user_id, created_at DESC)
    WHERE user_id IS NOT NULL;

-- route is low-cardinality (one row per defined endpoint) — safe to index.
CREATE INDEX IF NOT EXISTS idx_http_req_logs_route
    ON http_request_logs (route, created_at DESC)
    WHERE route IS NOT NULL;

-- Bootstrap partitions: current month + 2 months ahead. A cron job (or
-- pg_partman) MUST create future partitions before they are needed.
DO $$
DECLARE
    d date := date_trunc('month', CURRENT_DATE)::date;
    i int;
BEGIN
    FOR i IN 0..2 LOOP
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS http_request_logs_%s PARTITION OF http_request_logs FOR VALUES FROM (%L) TO (%L)',
            to_char(d + (i || ' months')::interval, 'YYYY_MM'),
            d + (i || ' months')::interval,
            d + ((i + 1) || ' months')::interval
        );
    END LOOP;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS http_request_logs CASCADE;
-- +goose StatementEnd
