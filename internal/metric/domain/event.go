package domain

import (
	"time"

	"github.com/google/uuid"
)

// FunctionMetricRecorded is raised when a new function metric is recorded.
type FunctionMetricRecorded struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	Name        string
	LatencyMs   float64
	IsPanic     bool
}

func NewFunctionMetricRecorded(id uuid.UUID, name string, latencyMs float64, isPanic bool) FunctionMetricRecorded {
	return FunctionMetricRecorded{
		aggregateID: id,
		occurredAt:  time.Now(),
		Name:        name,
		LatencyMs:   latencyMs,
		IsPanic:     isPanic,
	}
}

func (e FunctionMetricRecorded) EventName() string      { return "metric.recorded" }
func (e FunctionMetricRecorded) OccurredAt() time.Time   { return e.occurredAt }
func (e FunctionMetricRecorded) AggregateID() uuid.UUID  { return e.aggregateID }
