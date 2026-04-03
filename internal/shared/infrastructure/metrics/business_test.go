package metrics

import (
	"context"
	"sync"
	"testing"
)

func TestNewBusinessMetrics(t *testing.T) {
	bm := NewBusinessMetrics("test-service")
	if bm == nil {
		t.Fatal("expected non-nil BusinessMetrics")
	}
	if bm.counters == nil {
		t.Fatal("expected non-nil counters map")
	}
	if len(bm.counters) != 0 {
		t.Fatalf("expected empty counters map, got %d entries", len(bm.counters))
	}
}

func TestBusinessMetrics_Inc(t *testing.T) {
	bm := NewBusinessMetrics("test-service")
	ctx := context.Background()

	// Should not panic on first call.
	bm.Inc(ctx, "orders_created")

	bm.mu.RLock()
	_, ok := bm.counters["orders_created"]
	bm.mu.RUnlock()

	if !ok {
		t.Fatal("expected counter to be created after first Inc call")
	}
}

func TestBusinessMetrics_Inc_NilReceiver(t *testing.T) {
	var bm *BusinessMetrics
	// Must not panic when receiver is nil.
	bm.Inc(context.Background(), "should_not_panic")
}

func TestBusinessMetrics_Inc_SameName(t *testing.T) {
	bm := NewBusinessMetrics("test-service")
	ctx := context.Background()

	bm.Inc(ctx, "logins")
	bm.Inc(ctx, "logins")

	bm.mu.RLock()
	count := len(bm.counters)
	bm.mu.RUnlock()

	if count != 1 {
		t.Fatalf("expected 1 counter entry for repeated name, got %d", count)
	}
}

func TestBusinessMetrics_Inc_DifferentNames(t *testing.T) {
	bm := NewBusinessMetrics("test-service")
	ctx := context.Background()

	bm.Inc(ctx, "alpha")
	bm.Inc(ctx, "beta")

	bm.mu.RLock()
	count := len(bm.counters)
	bm.mu.RUnlock()

	if count != 2 {
		t.Fatalf("expected 2 counter entries for different names, got %d", count)
	}
}

func TestBusinessMetrics_ConcurrentAccess(t *testing.T) {
	bm := NewBusinessMetrics("test-service")
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			bm.Inc(ctx, "concurrent_counter")
		}(i)
	}
	wg.Wait()

	bm.mu.RLock()
	_, ok := bm.counters["concurrent_counter"]
	bm.mu.RUnlock()

	if !ok {
		t.Fatal("expected counter to exist after concurrent Inc calls")
	}
}
