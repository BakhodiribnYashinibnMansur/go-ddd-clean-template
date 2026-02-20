package logger

import (
	"context"
)

// safeContext returns the original context if it's not cancelled, otherwise returns background context
// This ensures logging always works even if the context is cancelled
func safeContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	// Check if context is cancelled
	select {
	case <-ctx.Done():
		// Context is cancelled, create a new background context but preserve values
		return context.WithoutCancel(ctx)
	default:
		// Context is still active
		return ctx
	}
}

// Implementation of context-aware methods
func (l *logger) Debugc(ctx context.Context, msg string, keysAndValues ...any) {
	ctx = safeContext(ctx)
	fields := extractFields(ctx)
	l.ctxZap.Debugw(msg, mergeFields(fields, keysAndValues...)...)
}

func (l *logger) Infoc(ctx context.Context, msg string, keysAndValues ...any) {
	ctx = safeContext(ctx)
	fields := extractFields(ctx)
	l.ctxZap.Infow(msg, mergeFields(fields, keysAndValues...)...)
}

func (l *logger) Warnc(ctx context.Context, msg string, keysAndValues ...any) {
	ctx = safeContext(ctx)
	fields := extractFields(ctx)
	l.ctxZap.Warnw(msg, mergeFields(fields, keysAndValues...)...)
}

func (l *logger) Errorc(ctx context.Context, msg string, keysAndValues ...any) {
	ctx = safeContext(ctx)
	fields := extractFields(ctx)
	l.ctxZap.Errorw(msg, mergeFields(fields, keysAndValues...)...)
}

func (l *logger) Fatalc(ctx context.Context, msg string, keysAndValues ...any) {
	ctx = safeContext(ctx)
	fields := extractFields(ctx)
	l.ctxZap.Fatalw(msg, mergeFields(fields, keysAndValues...)...)
}
