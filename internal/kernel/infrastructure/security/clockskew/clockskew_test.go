package clockskew_test

import (
	"testing"
	"time"

	"gct/internal/kernel/infrastructure/security/clockskew"
)

func TestObserveDoesNotPanicForZeroTime(t *testing.T) {
	// Should not panic even with a zero-value time.
	clockskew.Observe("test", time.Time{})
}

func TestObserveWithRecentIAT(t *testing.T) {
	// Observing a very recent issued-at should not panic and should record
	// a small value. We cannot easily inspect the histogram value without
	// pulling in testutil, so we just verify no panic.
	clockskew.Observe("test", time.Now().Add(-1*time.Second))
}
