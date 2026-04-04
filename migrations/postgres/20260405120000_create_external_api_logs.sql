-- +goose Up
-- +goose StatementBegin

-- external_api_logs stores 3rd-party HTTP call failures. Partitioned monthly:
-- retention = DROP PARTITION (O(1)) rather than DELETE (rewrite heap).
CREATE TABLE IF NOT EXISTS external_api_logs (
    id UUID NOT NULL DEFAULT gen_random_uuid(),

    api_name            VARCHAR(64)  NOT NULL,
    operation           VARCHAR(128),

    request_method      VARCHAR(10)  NOT NULL,
    request_url         TEXT         NOT NULL,
    request_headers     TEXT,
    request_body        TEXT,
    request_body_size   INT,

    response_status     INT,
    response_headers    TEXT,
    response_body       TEXT,
    response_body_size  INT,

    error_text          TEXT,
    duration_ms         INT          NOT NULL,

    request_id          VARCHAR(64),
    user_id             VARCHAR(128),
    session_id          VARCHAR(64),
    ip_address          VARCHAR(45),

    created_at          TIMESTAMP    NOT NULL DEFAULT NOW(),

    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE INDEX IF NOT EXISTS idx_ext_api_logs_api_name
    ON external_api_logs (api_name, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_ext_api_logs_status
    ON external_api_logs (response_status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_ext_api_logs_request_id
    ON external_api_logs (request_id)
    WHERE request_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_ext_api_logs_operation
    ON external_api_logs (api_name, operation, created_at DESC)
    WHERE operation IS NOT NULL;

-- Bootstrap partitions: current month + 2 ahead.
DO $$
DECLARE
    d date := date_trunc('month', CURRENT_DATE)::date;
    i int;
BEGIN
    FOR i IN 0..2 LOOP
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS external_api_logs_%s PARTITION OF external_api_logs FOR VALUES FROM (%L) TO (%L)',
            to_char(d + (i || ' months')::interval, 'YYYY_MM'),
            d + (i || ' months')::interval,
            d + ((i + 1) || ' months')::interval
        );
    END LOOP;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS external_api_logs CASCADE;
-- +goose StatementEnd
