package http

// CreateRequest represents the request body for recording a metric.
type CreateRequest struct {
	Name       string  `json:"name" binding:"required"`
	LatencyMs  float64 `json:"latency_ms" binding:"required"`
	IsPanic    bool    `json:"is_panic"`
	PanicError *string `json:"panic_error,omitempty"`
}
