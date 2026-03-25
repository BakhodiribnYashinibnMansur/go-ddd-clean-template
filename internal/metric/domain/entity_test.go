package domain_test

import (
	"testing"
	"time"

	domain "gct/internal/metric/domain"

	"github.com/google/uuid"
)

func TestNewFunctionMetric(t *testing.T) {
	fm := domain.NewFunctionMetric("UserService.Create", 150.5, false, nil)

	if fm.Name() != "UserService.Create" {
		t.Fatalf("expected name UserService.Create, got %s", fm.Name())
	}
	if fm.LatencyMs() != 150.5 {
		t.Fatalf("expected latencyMs 150.5, got %f", fm.LatencyMs())
	}
	if fm.IsPanic() {
		t.Fatal("isPanic should be false")
	}
	if fm.PanicError() != nil {
		t.Fatal("panicError should be nil")
	}

	// Should have a FunctionMetricRecorded event.
	events := fm.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventName() != "metric.recorded" {
		t.Fatalf("expected metric.recorded, got %s", events[0].EventName())
	}
}

func TestNewFunctionMetric_WithPanic(t *testing.T) {
	panicErr := "runtime error: index out of range"
	fm := domain.NewFunctionMetric("Handler.Process", 0.5, true, &panicErr)

	if !fm.IsPanic() {
		t.Fatal("isPanic should be true")
	}
	if fm.PanicError() == nil || *fm.PanicError() != panicErr {
		t.Fatal("panicError should match")
	}
}

func TestReconstructFunctionMetric(t *testing.T) {
	id := uuid.New()
	createdAt := time.Now()
	panicErr := "nil pointer"

	fm := domain.ReconstructFunctionMetric(id, createdAt, "TestFunc", 42.0, true, &panicErr)

	if fm.ID() != id {
		t.Fatal("ID mismatch")
	}
	if fm.Name() != "TestFunc" {
		t.Fatal("name mismatch")
	}
	if fm.LatencyMs() != 42.0 {
		t.Fatal("latencyMs mismatch")
	}
	if !fm.IsPanic() {
		t.Fatal("isPanic mismatch")
	}
	if fm.PanicError() == nil || *fm.PanicError() != panicErr {
		t.Fatal("panicError mismatch")
	}

	// Reconstruct should not raise events.
	if len(fm.Events()) != 0 {
		t.Fatalf("expected 0 events, got %d", len(fm.Events()))
	}
}
