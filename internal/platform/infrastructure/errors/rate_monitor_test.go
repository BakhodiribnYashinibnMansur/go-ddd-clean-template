package errors_test

import (
	"testing"
	"time"

	apperrors "gct/internal/platform/infrastructure/errors"
)

func TestRateMonitor_DetectsThresholdBreach(t *testing.T) {
	var alertedCode string
	monitor := apperrors.NewRateMonitor(apperrors.RateMonitorConfig{
		Window:    1 * time.Second,
		Threshold: 3,
		OnBreach: func(code string, count int) {
			alertedCode = code
		},
	})

	for i := 0; i < 4; i++ {
		monitor.Record("TEST_ERROR")
	}

	if alertedCode != "TEST_ERROR" {
		t.Fatalf("expected breach alert for TEST_ERROR, got %q", alertedCode)
	}
}

func TestRateMonitor_NoBreach_UnderThreshold(t *testing.T) {
	breached := false
	monitor := apperrors.NewRateMonitor(apperrors.RateMonitorConfig{
		Window:    1 * time.Second,
		Threshold: 10,
		OnBreach: func(code string, count int) {
			breached = true
		},
	})

	for i := 0; i < 5; i++ {
		monitor.Record("TEST_ERROR")
	}

	if breached {
		t.Fatal("should not breach at 5/10 threshold")
	}
}

func TestRateMonitor_WindowExpiry(t *testing.T) {
	breachCount := 0
	monitor := apperrors.NewRateMonitor(apperrors.RateMonitorConfig{
		Window:    50 * time.Millisecond,
		Threshold: 3,
		OnBreach: func(code string, count int) {
			breachCount++
		},
	})

	monitor.Record("TEST")
	monitor.Record("TEST")
	monitor.Record("TEST")
	time.Sleep(60 * time.Millisecond)

	monitor.Record("TEST")
	monitor.Record("TEST")
	monitor.Record("TEST")
	monitor.Record("TEST")

	if breachCount < 1 {
		t.Fatalf("expected at least 1 breach, got %d", breachCount)
	}
}
