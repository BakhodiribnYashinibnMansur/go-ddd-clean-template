package event_test

import (
	"testing"
	"time"

	"gct/internal/context/ops/generic/metric/domain/event"

	"github.com/google/uuid"
)

func TestNewFunctionMetricRecorded(t *testing.T) {
	id := uuid.New()
	before := time.Now()

	e := event.NewFunctionMetricRecorded(id, "UserService.Create", 150.5, false)

	if e.EventName() != "metric.recorded" {
		t.Fatalf("expected metric.recorded, got %s", e.EventName())
	}
	if e.AggregateID() != id {
		t.Fatalf("expected aggregate ID %s, got %s", id, e.AggregateID())
	}
	if e.Name != "UserService.Create" {
		t.Fatalf("expected name UserService.Create, got %s", e.Name)
	}
	if e.LatencyMs != 150.5 {
		t.Fatalf("expected latencyMs 150.5, got %f", e.LatencyMs)
	}
	if e.IsPanic {
		t.Fatal("expected isPanic false")
	}
	if e.OccurredAt().Before(before) {
		t.Fatal("occurredAt should be >= test start time")
	}
}

func TestNewFunctionMetricRecorded_WithPanic(t *testing.T) {
	id := uuid.New()

	e := event.NewFunctionMetricRecorded(id, "Handler.Process", 0.5, true)

	if !e.IsPanic {
		t.Fatal("expected isPanic true")
	}
	if e.Name != "Handler.Process" {
		t.Fatalf("expected Handler.Process, got %s", e.Name)
	}
}
