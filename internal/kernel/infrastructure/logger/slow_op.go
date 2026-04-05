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

// SlowWarnLogger is the narrow subset of Log that SlowOp needs. Accepting this
// narrow interface lets callers pass handlers that depend only on the methods
// they actually use (ISP), while any value satisfying Log still satisfies this.
type SlowWarnLogger interface {
	Warnc(ctx context.Context, msg string, keysAndValues ...any)
}

// SlowOp returns a function that, when called in a defer, logs a warning if the
// operation exceeded the slow threshold. Designed to pair with pgxutil.AppSpan:
//
//	ctx, end := pgxutil.AppSpan(ctx, "CreateAnnouncementHandler.Handle")
//	defer func() { end(err) }()
//	defer logger.SlowOp(h.logger, ctx, "CreateAnnouncement", "announcement")()
func SlowOp(l SlowWarnLogger, ctx context.Context, op, entity string) func() {
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
