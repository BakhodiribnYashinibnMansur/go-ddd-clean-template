package domain

import (
	"time"

	"github.com/google/uuid"
)

type FunctionMetric struct {
	ID         uuid.UUID `db:"id"          json:"id"`
	Name       string    `db:"name"        json:"name"`
	LatencyMs  int       `db:"latency_ms"  json:"latency_ms"`
	IsPanic    bool      `db:"is_panic"    json:"is_panic"`
	PanicError *string   `db:"panic_error" json:"panic_error,omitempty"`
	CreatedAt  time.Time `db:"created_at"  json:"created_at"`
}

type FunctionMetricsFilter struct {
	Name       *string     `json:"name,omitempty"`
	IsPanic    *bool       `json:"is_panic,omitempty"`
	FromDate   *time.Time  `json:"from_date,omitempty"`
	ToDate     *time.Time  `json:"to_date,omitempty"`
	Pagination *Pagination `json:"pagination"`
}
