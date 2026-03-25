package domain_test

import (
	"testing"
	"time"

	domain "gct/internal/ratelimit/domain"

	"github.com/google/uuid"
)

func TestNewRateLimit(t *testing.T) {
	rl := domain.NewRateLimit("api_global", "/api/*", 100, 60, true)

	if rl.Name() != "api_global" {
		t.Fatalf("expected name api_global, got %s", rl.Name())
	}
	if rl.Rule() != "/api/*" {
		t.Fatalf("expected rule /api/*, got %s", rl.Rule())
	}
	if rl.RequestsPerWindow() != 100 {
		t.Fatalf("expected requestsPerWindow 100, got %d", rl.RequestsPerWindow())
	}
	if rl.WindowDuration() != 60 {
		t.Fatalf("expected windowDuration 60, got %d", rl.WindowDuration())
	}
	if !rl.Enabled() {
		t.Fatal("expected enabled true")
	}
}

func TestRateLimit_Update(t *testing.T) {
	rl := domain.NewRateLimit("api_global", "/api/*", 100, 60, true)

	newRequests := 200
	rl.Update(nil, nil, &newRequests, nil, nil)

	if rl.RequestsPerWindow() != 200 {
		t.Fatalf("expected requestsPerWindow 200, got %d", rl.RequestsPerWindow())
	}

	events := rl.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventName() != "ratelimit.changed" {
		t.Fatalf("expected ratelimit.changed, got %s", events[0].EventName())
	}
}

func TestReconstructRateLimit(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	rl := domain.ReconstructRateLimit(id, now, now, "test", "/test", 50, 30, false)

	if rl.ID() != id {
		t.Fatal("ID mismatch")
	}
	if rl.Name() != "test" {
		t.Fatal("name mismatch")
	}
	if rl.Enabled() {
		t.Fatal("should not be enabled")
	}
	if len(rl.Events()) != 0 {
		t.Fatalf("expected 0 events, got %d", len(rl.Events()))
	}
}
