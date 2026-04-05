package errorx_test

import (
	"testing"
	"time"

	apperrors "gct/internal/kernel/infrastructure/errorx"
)

func TestSLOTracker_SuccessRate(t *testing.T) {
	tracker := apperrors.NewSLOTracker(apperrors.SLOConfig{
		Target: 0.99, // 99%
		Window: time.Hour,
	})

	for i := 0; i < 99; i++ {
		tracker.RecordSuccess()
	}
	tracker.RecordError()

	stats := tracker.Stats()
	if stats.TotalRequests != 100 {
		t.Fatalf("expected 100 total, got %d", stats.TotalRequests)
	}
	if stats.ErrorRequests != 1 {
		t.Fatalf("expected 1 error, got %d", stats.ErrorRequests)
	}
	if stats.SuccessRate < 0.98 || stats.SuccessRate > 1.0 {
		t.Fatalf("expected ~0.99 success rate, got %f", stats.SuccessRate)
	}
}

func TestSLOTracker_BudgetExhausted(t *testing.T) {
	exhausted := false
	tracker := apperrors.NewSLOTracker(apperrors.SLOConfig{
		Target: 0.99,
		Window: time.Hour,
		OnBudgetExhausted: func(stats apperrors.SLOStats) {
			exhausted = true
		},
	})

	// 100 success + 10 errors = 90.9% (below 99% target)
	for i := 0; i < 100; i++ {
		tracker.RecordSuccess()
	}
	for i := 0; i < 10; i++ {
		tracker.RecordError()
	}

	if !exhausted {
		t.Fatal("expected budget exhausted callback")
	}

	stats := tracker.Stats()
	if !stats.Exhausted {
		t.Fatal("expected Exhausted=true")
	}
}

func TestSLOTracker_WindowReset(t *testing.T) {
	tracker := apperrors.NewSLOTracker(apperrors.SLOConfig{
		Target: 0.99,
		Window: 50 * time.Millisecond,
	})

	tracker.RecordError()
	time.Sleep(60 * time.Millisecond)
	tracker.RecordSuccess() // triggers window reset

	stats := tracker.Stats()
	if stats.ErrorRequests != 0 {
		t.Fatalf("expected 0 errors after window reset, got %d", stats.ErrorRequests)
	}
}
