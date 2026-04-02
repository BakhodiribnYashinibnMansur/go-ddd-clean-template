package latency

import (
	"context"
	"time"

	"gct/internal/shared/infrastructure/logger"
)

// Reporter periodically logs latency stats and checks alerts.
type Reporter struct {
	tracker  *Tracker
	alert    *AlertManager
	interval time.Duration
	window   time.Duration
	log      logger.Log
	cancel   context.CancelFunc
}

// NewReporter creates a new periodic reporter.
func NewReporter(tracker *Tracker, alert *AlertManager, interval time.Duration, windowSec int, log logger.Log) *Reporter {
	return &Reporter{
		tracker:  tracker,
		alert:    alert,
		interval: interval,
		window:   time.Duration(windowSec) * time.Second,
		log:      log,
	}
}

// Start begins the periodic logging and alerting loop.
func (r *Reporter) Start(ctx context.Context) {
	ctx, r.cancel = context.WithCancel(ctx)
	go r.run(ctx)
}

// Stop cancels the reporter goroutine.
func (r *Reporter) Stop() {
	if r.cancel != nil {
		r.cancel()
	}
}

func (r *Reporter) run(ctx context.Context) {
	logTicker := time.NewTicker(r.interval)
	windowTicker := time.NewTicker(r.window)
	defer logTicker.Stop()
	defer windowTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-logTicker.C:
			stats := r.tracker.Stats()
			if stats.Count > 0 {
				r.log.Infow("latency stats",
					"p50", stats.P50.String(),
					"p95", stats.P95.String(),
					"p99", stats.P99.String(),
					"mean", stats.Mean.String(),
					"count", stats.Count,
				)
			}
			if r.alert != nil {
				r.alert.Check(stats)
			}
		case <-windowTicker.C:
			r.tracker.Reset()
			r.log.Infow("latency tracker window reset")
		}
	}
}
