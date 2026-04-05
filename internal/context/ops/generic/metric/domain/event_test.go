package domain_test

import (
	"testing"
	"time"

	domain "gct/internal/context/ops/generic/metric/domain"
)

func TestNewFunctionMetricRecorded(t *testing.T) {
	id := domain.NewMetricID()
	before := time.Now()

	event := domain.NewFunctionMetricRecorded(id.UUID(), "UserService.Create", 150.5, false)

	if event.EventName() != "metric.recorded" {
		t.Fatalf("expected metric.recorded, got %s", event.EventName())
	}
	if event.AggregateID() != id.UUID() {
		t.Fatalf("expected aggregate ID %s, got %s", id, event.AggregateID())
	}
	if event.Name != "UserService.Create" {
		t.Fatalf("expected name UserService.Create, got %s", event.Name)
	}
	if event.LatencyMs != 150.5 {
		t.Fatalf("expected latencyMs 150.5, got %f", event.LatencyMs)
	}
	if event.IsPanic {
		t.Fatal("expected isPanic false")
	}
	if event.OccurredAt().Before(before) {
		t.Fatal("occurredAt should be >= test start time")
	}
}

func TestNewFunctionMetricRecorded_WithPanic(t *testing.T) {
	id := domain.NewMetricID()

	event := domain.NewFunctionMetricRecorded(id.UUID(), "Handler.Process", 0.5, true)

	if !event.IsPanic {
		t.Fatal("expected isPanic true")
	}
	if event.Name != "Handler.Process" {
		t.Fatalf("expected Handler.Process, got %s", event.Name)
	}
}
