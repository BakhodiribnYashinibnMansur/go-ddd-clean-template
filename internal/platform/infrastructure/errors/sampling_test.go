package errors_test

import (
	"testing"
	"time"

	apperrors "gct/internal/platform/infrastructure/errors"
)

func TestSampler_LogsAllUnderThreshold(t *testing.T) {
	s := apperrors.NewSampler(apperrors.SamplerConfig{
		Window:     time.Second,
		Threshold:  5,
		SampleRate: 10,
	})

	for i := 0; i < 5; i++ {
		if !s.ShouldLog("TEST") {
			t.Fatalf("should log all under threshold, failed at %d", i)
		}
	}
}

func TestSampler_SamplesOverThreshold(t *testing.T) {
	s := apperrors.NewSampler(apperrors.SamplerConfig{
		Window:     time.Second,
		Threshold:  5,
		SampleRate: 2, // log every 2nd
	})

	logged := 0
	for i := 0; i < 20; i++ {
		if s.ShouldLog("TEST") {
			logged++
		}
	}

	// 5 under threshold + ~7 sampled (every 2nd of remaining 15) ≈ 12
	if logged < 5 || logged > 15 {
		t.Fatalf("expected between 5-15 logged, got %d", logged)
	}
}
