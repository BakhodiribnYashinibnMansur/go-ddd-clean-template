package httpclient

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// OTel counters for the external-api-log pipeline. Gracefully no-op if the
// meter provider has not been initialised.
// The inc* helpers below use context.Background() because they are invoked from
// worker-goroutine code paths (sink/flusher) where no caller context is
// available and counter increments must not be tied to any request lifetime.
var (
	mDropped  metric.Int64Counter
	mFlushed  metric.Int64Counter
	mFailed   metric.Int64Counter
	mDupRisk  metric.Int64Counter
	mPoisoned metric.Int64Counter
	mEmitted  metric.Int64Counter // errors surfaced to the sink

	metricsReady bool
)

func initMetrics() {
	if metricsReady {
		return
	}
	m := otel.Meter("httpclient")
	var err error
	mDropped, err = m.Int64Counter("httpclient.entries.dropped",
		metric.WithDescription("Entries dropped due to in-memory buffer pressure"),
		metric.WithUnit("{entry}"))
	if err != nil {
		return
	}
	mFlushed, err = m.Int64Counter("httpclient.entries.flushed",
		metric.WithDescription("Entries persisted to PostgreSQL"),
		metric.WithUnit("{entry}"))
	if err != nil {
		return
	}
	mFailed, err = m.Int64Counter("httpclient.flush.failed",
		metric.WithDescription("Entries that failed to COPY FROM (will retry)"),
		metric.WithUnit("{entry}"))
	if err != nil {
		return
	}
	mDupRisk, err = m.Int64Counter("httpclient.flush.dup_risk",
		metric.WithDescription("Times LTrim failed after COPY, risking duplicates"),
		metric.WithUnit("{event}"))
	if err != nil {
		return
	}
	mPoisoned, err = m.Int64Counter("httpclient.entries.poisoned",
		metric.WithDescription("Malformed entries discarded"),
		metric.WithUnit("{entry}"))
	if err != nil {
		return
	}
	mEmitted, err = m.Int64Counter("httpclient.errors.emitted",
		metric.WithDescription("External-API errors captured and sent to sink"),
		metric.WithUnit("{error}"))
	if err != nil {
		return
	}
	metricsReady = true
}

func incDropped() {
	initMetrics()
	if mDropped != nil {
		mDropped.Add(context.Background(), 1)
	}
}
func incFlushed(n int) {
	initMetrics()
	if mFlushed != nil {
		mFlushed.Add(context.Background(), int64(n))
	}
}
func incFlushFailed(n int) {
	initMetrics()
	if mFailed != nil {
		mFailed.Add(context.Background(), int64(n))
	}
}
func incFlushDupRisk() {
	initMetrics()
	if mDupRisk != nil {
		mDupRisk.Add(context.Background(), 1)
	}
}
func incFlushPoisoned() {
	initMetrics()
	if mPoisoned != nil {
		mPoisoned.Add(context.Background(), 1)
	}
}
func incEmitted() {
	initMetrics()
	if mEmitted != nil {
		mEmitted.Add(context.Background(), 1)
	}
}
