package errors

import (
	"context"
	"sync/atomic"
	"time"
)

// SLOConfig configures the SLO tracker.
type SLOConfig struct {
	Target            float64       // target success rate (e.g., 0.999 = 99.9%)
	Window            time.Duration // measurement window (e.g., 1 hour)
	OnBudgetExhausted func(stats SLOStats) // called when error budget is exhausted
}

// SLOStats holds current SLO statistics.
type SLOStats struct {
	TotalRequests   int64   `json:"total_requests"`
	ErrorRequests   int64   `json:"error_requests"`
	SuccessRate     float64 `json:"success_rate"`
	Target          float64 `json:"target"`
	BudgetTotal     int64   `json:"budget_total"`    // max allowed errors
	BudgetUsed      int64   `json:"budget_used"`     // errors so far
	BudgetRemaining int64   `json:"budget_remaining"`
	Exhausted       bool    `json:"exhausted"`
	WindowStart     string  `json:"window_start"`
}

// SLOTracker tracks error budget based on success rate target.
type SLOTracker struct {
	cfg         SLOConfig
	total       atomic.Int64
	errors      atomic.Int64
	windowStart atomic.Int64 // unix timestamp
	alerted     atomic.Bool
}

// NewSLOTracker creates a new SLO tracker.
func NewSLOTracker(cfg SLOConfig) *SLOTracker {
	if cfg.Target <= 0 {
		cfg.Target = 0.999
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Hour
	}
	s := &SLOTracker{cfg: cfg}
	s.windowStart.Store(time.Now().Unix())
	return s
}

// RecordSuccess records a successful request.
func (s *SLOTracker) RecordSuccess() {
	s.maybeResetWindow()
	s.total.Add(1)
}

// RecordError records a failed request.
func (s *SLOTracker) RecordError() {
	s.maybeResetWindow()
	s.total.Add(1)
	s.errors.Add(1)
	s.checkBudget()
}

// Stats returns current SLO statistics.
func (s *SLOTracker) Stats() SLOStats {
	total := s.total.Load()
	errs := s.errors.Load()

	var successRate float64
	if total > 0 {
		successRate = float64(total-errs) / float64(total)
	} else {
		successRate = 1.0
	}

	budgetTotal := int64(float64(total) * (1 - s.cfg.Target))
	if budgetTotal < 1 && total > 0 {
		budgetTotal = 1
	}
	budgetRemaining := budgetTotal - errs
	if budgetRemaining < 0 {
		budgetRemaining = 0
	}

	return SLOStats{
		TotalRequests:   total,
		ErrorRequests:   errs,
		SuccessRate:     successRate,
		Target:          s.cfg.Target,
		BudgetTotal:     budgetTotal,
		BudgetUsed:      errs,
		BudgetRemaining: budgetRemaining,
		Exhausted:       errs > budgetTotal && total > 100,
		WindowStart:     time.Unix(s.windowStart.Load(), 0).UTC().Format(time.RFC3339),
	}
}

func (s *SLOTracker) maybeResetWindow() {
	start := time.Unix(s.windowStart.Load(), 0)
	if time.Since(start) > s.cfg.Window {
		s.total.Store(0)
		s.errors.Store(0)
		s.windowStart.Store(time.Now().Unix())
		s.alerted.Store(false)
	}
}

func (s *SLOTracker) checkBudget() {
	stats := s.Stats()
	if stats.Exhausted && !s.alerted.Load() {
		s.alerted.Store(true)
		if s.cfg.OnBudgetExhausted != nil {
			s.cfg.OnBudgetExhausted(stats)
		}
	}
}

// SLOMiddlewareHook returns an ErrorHook that records errors for SLO tracking.
func SLOMiddlewareHook(tracker *SLOTracker) func(ctx context.Context, err *AppError) {
	return func(ctx context.Context, err *AppError) {
		tracker.RecordError()
	}
}
