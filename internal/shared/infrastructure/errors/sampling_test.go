package errors

import (
	"sync"
	"testing"
	"time"
)

func TestNewSampler_Defaults(t *testing.T) {
	s := NewSampler(SamplerConfig{})

	if s.cfg.Window != time.Minute {
		t.Errorf("Window: got %v, want %v", s.cfg.Window, time.Minute)
	}
	if s.cfg.Threshold != 100 {
		t.Errorf("Threshold: got %d, want 100", s.cfg.Threshold)
	}
	if s.cfg.SampleRate != 10 {
		t.Errorf("SampleRate: got %d, want 10", s.cfg.SampleRate)
	}
}

func TestNewSampler_NegativeValuesGetDefaults(t *testing.T) {
	s := NewSampler(SamplerConfig{
		Window:     -5 * time.Second,
		Threshold:  -1,
		SampleRate: -1,
	})

	if s.cfg.Window != time.Minute {
		t.Errorf("Window: got %v, want %v", s.cfg.Window, time.Minute)
	}
	if s.cfg.Threshold != 100 {
		t.Errorf("Threshold: got %d, want 100", s.cfg.Threshold)
	}
	if s.cfg.SampleRate != 10 {
		t.Errorf("SampleRate: got %d, want 10", s.cfg.SampleRate)
	}
}

func TestNewSampler_CustomConfig(t *testing.T) {
	cfg := SamplerConfig{
		Window:     30 * time.Second,
		Threshold:  50,
		SampleRate: 5,
	}
	s := NewSampler(cfg)

	if s.cfg.Window != 30*time.Second {
		t.Errorf("Window: got %v, want %v", s.cfg.Window, 30*time.Second)
	}
	if s.cfg.Threshold != 50 {
		t.Errorf("Threshold: got %d, want 50", s.cfg.Threshold)
	}
	if s.cfg.SampleRate != 5 {
		t.Errorf("SampleRate: got %d, want 5", s.cfg.SampleRate)
	}
}

func TestShouldLog_UnderThreshold(t *testing.T) {
	s := NewSampler(SamplerConfig{
		Window:     time.Minute,
		Threshold:  10,
		SampleRate: 5,
	})

	for i := 1; i <= 10; i++ {
		if !s.ShouldLog("ERR_TEST") {
			t.Fatalf("call %d: expected true (under threshold), got false", i)
		}
	}
}

func TestShouldLog_OverThreshold_SamplesEveryNth(t *testing.T) {
	const threshold = 5
	const sampleRate = 3

	s := NewSampler(SamplerConfig{
		Window:     time.Minute,
		Threshold:  threshold,
		SampleRate: sampleRate,
	})

	// Exhaust the threshold (calls 1..5 all return true).
	for i := 1; i <= threshold; i++ {
		s.ShouldLog("CODE")
	}

	// After threshold, only every sampleRate-th total count should log.
	// count starts at 6 now.
	tests := []struct {
		callNum int  // the overall count after this call
		want    bool // count % sampleRate == 0
	}{
		{6, 6%sampleRate == 0},   // 6%3==0 -> true
		{7, 7%sampleRate == 0},   // false
		{8, 8%sampleRate == 0},   // false
		{9, 9%sampleRate == 0},   // 9%3==0 -> true
		{10, 10%sampleRate == 0}, // false
		{11, 11%sampleRate == 0}, // false
		{12, 12%sampleRate == 0}, // 12%3==0 -> true
	}

	for _, tt := range tests {
		got := s.ShouldLog("CODE")
		if got != tt.want {
			t.Errorf("count=%d: got %v, want %v", tt.callNum, got, tt.want)
		}
	}
}

func TestShouldLog_WindowReset(t *testing.T) {
	s := NewSampler(SamplerConfig{
		Window:     50 * time.Millisecond,
		Threshold:  2,
		SampleRate: 100, // large so nothing passes after threshold
	})

	// Fill up the window.
	s.ShouldLog("RST")
	s.ShouldLog("RST")
	// Third call is over threshold; with sampleRate=100 it should be false.
	if s.ShouldLog("RST") {
		t.Fatal("expected false for over-threshold call before window reset")
	}

	// Wait for the window to expire.
	time.Sleep(60 * time.Millisecond)

	// After reset the counter restarts; first call in new window must be true.
	if !s.ShouldLog("RST") {
		t.Fatal("expected true after window reset")
	}
}

func TestShouldLog_DifferentCodesTrackedIndependently(t *testing.T) {
	s := NewSampler(SamplerConfig{
		Window:     time.Minute,
		Threshold:  2,
		SampleRate: 100,
	})

	// Push code-A past its threshold.
	s.ShouldLog("A")
	s.ShouldLog("A")
	if s.ShouldLog("A") {
		t.Fatal("code A: expected false after exceeding threshold")
	}

	// Code-B should still be under its own threshold.
	if !s.ShouldLog("B") {
		t.Fatal("code B: expected true (independent tracking)")
	}
	if !s.ShouldLog("B") {
		t.Fatal("code B: expected true (still under threshold)")
	}
}

func TestShouldLog_ConcurrentAccess(t *testing.T) {
	s := NewSampler(SamplerConfig{
		Window:     time.Minute,
		Threshold:  1000,
		SampleRate: 10,
	})

	const goroutines = 50
	const callsPerGoroutine = 200

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < callsPerGoroutine; i++ {
				_ = s.ShouldLog("CONCURRENT")
			}
		}()
	}
	wg.Wait()

	// Verify the total count is consistent.
	s.mu.Lock()
	w := s.windows["CONCURRENT"]
	s.mu.Unlock()

	total := int(w.count)
	want := goroutines * callsPerGoroutine
	if total != want {
		t.Errorf("total count: got %d, want %d", total, want)
	}
}

func BenchmarkShouldLog(b *testing.B) {
	s := NewSampler(SamplerConfig{
		Window:     time.Minute,
		Threshold:  100,
		SampleRate: 10,
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.ShouldLog("BENCH")
		}
	})
}
