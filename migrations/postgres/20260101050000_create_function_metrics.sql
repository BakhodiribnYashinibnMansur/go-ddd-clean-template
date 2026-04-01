-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS function_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    latency_ms INTEGER NOT NULL,
    is_panic BOOLEAN DEFAULT FALSE,
    panic_error TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_func_metrics_name ON function_metrics(name);
CREATE INDEX IF NOT EXISTS idx_func_metrics_created_at ON function_metrics(created_at);
CREATE INDEX IF NOT EXISTS idx_func_metrics_panic ON function_metrics(is_panic);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS function_metrics;
-- +goose StatementEnd
