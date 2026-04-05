package command

import "context"

// commandLogger is the minimal subset of logging methods that user-BC command
// handlers need. It follows the Go idiom of "accept interfaces": consumers
// declare exactly what they use, so any concrete logger with these methods
// (such as logger.Log) can be injected.
//
// The full kernel logger.Log interface automatically satisfies this.
type commandLogger interface {
	Errorc(ctx context.Context, msg string, keysAndValues ...any)
	Warnc(ctx context.Context, msg string, keysAndValues ...any)
}
