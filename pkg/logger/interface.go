package logger

import "context"

// Log is the main logger interface.
// It provides both context-aware methods and simple methods.
type Log interface {
	// Simple methods
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Debugw(msg string, keysAndValues ...interface{})
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatal(args ...interface{})
	Fatalf(template string, args ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})

	// Context-aware methods
	Debugc(ctx context.Context, msg string, keysAndValues ...interface{})
	Infoc(ctx context.Context, msg string, keysAndValues ...interface{})
	Warnc(ctx context.Context, msg string, keysAndValues ...interface{})
	Errorc(ctx context.Context, msg string, keysAndValues ...interface{})
	Fatalc(ctx context.Context, msg string, keysAndValues ...interface{})
}
