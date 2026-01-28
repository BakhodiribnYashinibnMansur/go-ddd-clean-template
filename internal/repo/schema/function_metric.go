package schema

// Table name
const TableFunctionMetric = "function_metrics"

// FunctionMetric table columns
const (
	FunctionMetricID         = "id"
	FunctionMetricName       = "name"
	FunctionMetricLatencyMs  = "latency_ms"
	FunctionMetricIsPanic    = "is_panic"
	FunctionMetricPanicError = "panic_error"
	FunctionMetricCreatedAt  = "created_at"
)
