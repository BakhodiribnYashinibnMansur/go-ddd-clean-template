package logger

import "context"

// Log is the main logger interface.
// It provides both context-aware methods and simple methods.
type Log interface {
	// Simple methods
	Debug(args ...any)
	Debugf(template string, args ...any)
	Debugw(msg string, keysAndValues ...any)
	Info(args ...any)
	Infof(template string, args ...any)
	Infow(msg string, keysAndValues ...any)
	Warn(args ...any)
	Warnf(template string, args ...any)
	Warnw(msg string, keysAndValues ...any)
	Error(args ...any)
	Errorf(template string, args ...any)
	Errorw(msg string, keysAndValues ...any)
	Fatal(args ...any)
	Fatalf(template string, args ...any)
	Fatalw(msg string, keysAndValues ...any)

	// Context-aware methods
	Debugc(ctx context.Context, msg string, keysAndValues ...any)
	Infoc(ctx context.Context, msg string, keysAndValues ...any)
	Warnc(ctx context.Context, msg string, keysAndValues ...any)
	Errorc(ctx context.Context, msg string, keysAndValues ...any)
	Fatalc(ctx context.Context, msg string, keysAndValues ...any)
}
