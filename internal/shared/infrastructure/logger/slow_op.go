package logger

import (
	"context"
	"time"
)

var slowOpThreshold = 500 * time.Millisecond

// SetSlowOpThreshold configures the global threshold for slow operation warnings.
func SetSlowOpThreshold(d time.Duration) {
	if d > 0 {
		slowOpThreshold = d
	}
}

// SlowOp returns a function that, when called in a defer, logs a warning if the
// operation exceeded the slow threshold. Designed to pair with pgxutil.AppSpan:
//
//	ctx, end := pgxutil.AppSpan(ctx, "CreateAnnouncementHandler.Handle")
//	defer func() { end(err) }()
//	defer logger.SlowOp(h.logger, ctx, "CreateAnnouncement", "announcement")()
func SlowOp(l Log, ctx context.Context, op, entity string) func() {
	start := time.Now()
	return func() {
		dur := time.Since(start)
		if dur >= slowOpThreshold {
			l.Warnc(ctx, "slow operation",
				"operation", op,
				"entity", entity,
				"duration_ms", dur.Milliseconds(),
				"threshold_ms", slowOpThreshold.Milliseconds(),
			)
		}
	}
}
