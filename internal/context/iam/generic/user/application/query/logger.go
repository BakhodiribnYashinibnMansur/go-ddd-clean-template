package query

import "context"

// queryLogger is the minimal subset of logging methods that user-BC query
// handlers need. Queries only emit warnings (slow-op traces, missing rows),
// never errors — errors are mapped to service-layer errors and returned.
//
// The full kernel logger.Log interface automatically satisfies this.
type queryLogger interface {
	Warnc(ctx context.Context, msg string, keysAndValues ...any)
}
