package reqlog

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// OTel metrics for the reqlog pipeline. All instruments are lazily initialised
// and gracefully degrade to no-ops if the meter provider has not been set up.
var (
	mDropped  metric.Int64Counter // entries dropped because the channel was full
	mFlushed  metric.Int64Counter // entries successfully COPY-FROM'd to PG
	mFailed   metric.Int64Counter // entries in a batch that failed to COPY FROM
	mDupRisk  metric.Int64Counter // times LTrim failed after a successful COPY
	mPoisoned metric.Int64Counter // entries discarded because they were malformed

	metricsReady bool
)

func initMetrics() {
	if metricsReady {
		return
	}
	m := otel.Meter("reqlog")
	var err error
	mDropped, err = m.Int64Counter("reqlog.entries.dropped",
		metric.WithDescription("Entries dropped due to in-memory buffer pressure"),
		metric.WithUnit("{entry}"))
	if err != nil {
		return
	}
	mFlushed, err = m.Int64Counter("reqlog.entries.flushed",
		metric.WithDescription("Entries persisted to PostgreSQL"),
		metric.WithUnit("{entry}"))
	if err != nil {
		return
	}
	mFailed, err = m.Int64Counter("reqlog.flush.failed",
		metric.WithDescription("Entries that failed to COPY FROM (will retry)"),
		metric.WithUnit("{entry}"))
	if err != nil {
		return
	}
	mDupRisk, err = m.Int64Counter("reqlog.flush.dup_risk",
		metric.WithDescription("Times LTrim failed after COPY, risking duplicates"),
		metric.WithUnit("{event}"))
	if err != nil {
		return
	}
	mPoisoned, err = m.Int64Counter("reqlog.entries.poisoned",
		metric.WithDescription("Malformed entries discarded"),
		metric.WithUnit("{entry}"))
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
