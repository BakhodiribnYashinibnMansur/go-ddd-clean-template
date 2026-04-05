package logger

import "context"

// BasicLogger emits messages at each log level without formatting.
type BasicLogger interface {
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Fatal(args ...any)
}

// FormatLogger emits printf-style formatted messages at each log level.
type FormatLogger interface {
	Debugf(template string, args ...any)
	Infof(template string, args ...any)
	Warnf(template string, args ...any)
	Errorf(template string, args ...any)
	Fatalf(template string, args ...any)
}

// StructuredLogger emits messages with key/value pairs at each log level.
type StructuredLogger interface {
	Debugw(msg string, keysAndValues ...any)
	Infow(msg string, keysAndValues ...any)
	Warnw(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
	Fatalw(msg string, keysAndValues ...any)
}

// ContextLogger emits context-aware messages (trace/request correlation).
type ContextLogger interface {
	Debugc(ctx context.Context, msg string, keysAndValues ...any)
	Infoc(ctx context.Context, msg string, keysAndValues ...any)
	Warnc(ctx context.Context, msg string, keysAndValues ...any)
	Errorc(ctx context.Context, msg string, keysAndValues ...any)
	Fatalc(ctx context.Context, msg string, keysAndValues ...any)
}

// Log is the full logger facade composed of the focused sub-interfaces.
// Consumers should prefer depending on the smallest sub-interface they need
// (e.g. StructuredLogger) rather than this full facade.
type Log interface {
	BasicLogger
	FormatLogger
	StructuredLogger
	ContextLogger
}
