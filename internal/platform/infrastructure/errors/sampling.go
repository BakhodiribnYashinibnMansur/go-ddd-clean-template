package errors

import (
	"sync"
	"sync/atomic"
	"time"
)

// SamplerConfig configures error log sampling.
type SamplerConfig struct {
	Window     time.Duration // sampling window (default: 1 minute)
	Threshold  int           // after this many errors per code, start sampling (default: 100)
	SampleRate int           // log every Nth error when sampling (default: 10)
}

// Sampler decides whether to log/report an error based on rate.
// Under threshold: log everything. Over threshold: log every Nth.
type Sampler struct {
	cfg     SamplerConfig
	mu      sync.Mutex
	windows map[string]*sampleWindow
}

type sampleWindow struct {
	count int64
	start time.Time
}

// NewSampler creates a new error sampler.
func NewSampler(cfg SamplerConfig) *Sampler {
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	if cfg.Threshold <= 0 {
		cfg.Threshold = 100
	}
	if cfg.SampleRate <= 0 {
		cfg.SampleRate = 10
	}
	return &Sampler{
		cfg:     cfg,
		windows: make(map[string]*sampleWindow),
	}
}

// ShouldLog returns true if this error should be logged.
func (s *Sampler) ShouldLog(code string) bool {
	s.mu.Lock()
	w, ok := s.windows[code]
	if !ok || time.Since(w.start) > s.cfg.Window {
		s.windows[code] = &sampleWindow{count: 1, start: time.Now()}
		s.mu.Unlock()
		return true
	}
	count := atomic.AddInt64(&w.count, 1)
	s.mu.Unlock()

	if int(count) <= s.cfg.Threshold {
		return true
	}
	// Sample: log every Nth
	return int(count)%s.cfg.SampleRate == 0
}
