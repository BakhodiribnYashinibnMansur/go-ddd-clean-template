package domain_test

import (
	"testing"
	"time"

	domain "gct/internal/context/ops/metric/domain"

	"github.com/google/uuid"
)

func TestNewFunctionMetricRecorded(t *testing.T) {
	id := uuid.New()
	before := time.Now()

	event := domain.NewFunctionMetricRecorded(id, "UserService.Create", 150.5, false)

	if event.EventName() != "metric.recorded" {
		t.Fatalf("expected metric.recorded, got %s", event.EventName())
	}
	if event.AggregateID() != id {
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
	id := uuid.New()

	event := domain.NewFunctionMetricRecorded(id, "Handler.Process", 0.5, true)

	if !event.IsPanic {
		t.Fatal("expected isPanic true")
	}
	if event.Name != "Handler.Process" {
		t.Fatalf("expected Handler.Process, got %s", event.Name)
	}
}
