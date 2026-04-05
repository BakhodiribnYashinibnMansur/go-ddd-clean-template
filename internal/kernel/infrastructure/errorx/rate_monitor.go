package errorx

import (
	"context"
	"sync"
	"time"
)

// RateMonitorConfig configures the error rate monitor.
type RateMonitorConfig struct {
	Window    time.Duration
	Threshold int
	OnBreach  func(code string, count int)
}

// RateMonitor tracks error rates per code using a sliding window.
type RateMonitor struct {
	cfg     RateMonitorConfig
	mu      sync.Mutex
	windows map[string]*rateWindow
}

type rateWindow struct {
	count    int
	start    time.Time
	breached bool
}

// NewRateMonitor creates a new rate monitor.
func NewRateMonitor(cfg RateMonitorConfig) *RateMonitor {
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	if cfg.Threshold <= 0 {
		cfg.Threshold = 10
	}
	return &RateMonitor{
		cfg:     cfg,
		windows: make(map[string]*rateWindow),
	}
}

// Record records an error occurrence and checks the threshold.
func (m *RateMonitor) Record(code string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	w, ok := m.windows[code]
	if !ok || time.Since(w.start) > m.cfg.Window {
		m.windows[code] = &rateWindow{count: 1, start: time.Now()}
		return
	}

	w.count++
	if w.count > m.cfg.Threshold && !w.breached {
		w.breached = true
		if m.cfg.OnBreach != nil {
			m.cfg.OnBreach(code, w.count)
		}
	}
}

// RateMonitorHook returns a function that feeds errors into the rate monitor.
func RateMonitorHook(monitor *RateMonitor) func(error) {
	return func(err error) {
		code := GetCode(err)
		if code != "" {
			monitor.Record(code)
		}
	}
}

// Cleanup removes expired windows to prevent memory leaks.
func (m *RateMonitor) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for code, w := range m.windows {
		if time.Since(w.start) > m.cfg.Window*2 {
			delete(m.windows, code)
		}
	}
}

// StartCleanup runs periodic cleanup in background. Call cancel to stop.
func (m *RateMonitor) StartCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(m.cfg.Window * 2)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.Cleanup()
			}
		}
	}()
}
