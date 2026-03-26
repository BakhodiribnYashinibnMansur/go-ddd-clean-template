package domain

import (
	"time"

	shared "gct/internal/shared/domain"

	"github.com/google/uuid"
)

// FunctionMetric is the aggregate root for recorded function-level performance metrics.
// It is append-only — metrics are never updated after creation, preserving an immutable time-series record.
// Panic information is captured separately to enable alerting on unrecovered panics.
type FunctionMetric struct {
	shared.AggregateRoot
	name       string
	latencyMs  float64
	isPanic    bool
	panicError *string
}

// NewFunctionMetric creates a new FunctionMetric aggregate and raises a FunctionMetricRecorded event.
func NewFunctionMetric(name string, latencyMs float64, isPanic bool, panicError *string) *FunctionMetric {
	fm := &FunctionMetric{
		AggregateRoot: shared.NewAggregateRoot(),
		name:          name,
		latencyMs:     latencyMs,
		isPanic:       isPanic,
		panicError:    panicError,
	}
	fm.AddEvent(NewFunctionMetricRecorded(fm.ID(), name, latencyMs, isPanic))
	return fm
}

// ReconstructFunctionMetric rebuilds a FunctionMetric aggregate from persisted data. No events are raised.
func ReconstructFunctionMetric(
	id uuid.UUID,
	createdAt time.Time,
	name string,
	latencyMs float64,
	isPanic bool,
	panicError *string,
) *FunctionMetric {
	return &FunctionMetric{
		AggregateRoot: shared.NewAggregateRootWithID(id, createdAt, createdAt, nil),
		name:          name,
		latencyMs:     latencyMs,
		isPanic:       isPanic,
		panicError:    panicError,
	}
}

// ---------------------------------------------------------------------------
// Getters
// ---------------------------------------------------------------------------

func (fm *FunctionMetric) Name() string        { return fm.name }
func (fm *FunctionMetric) LatencyMs() float64  { return fm.latencyMs }
func (fm *FunctionMetric) IsPanic() bool       { return fm.isPanic }
func (fm *FunctionMetric) PanicError() *string { return fm.panicError }
