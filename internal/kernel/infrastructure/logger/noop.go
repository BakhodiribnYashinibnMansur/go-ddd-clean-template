package logger

import "context"

type noopLogger struct{}

// Noop returns a logger that discards all output. Useful for tests.
func Noop() Log { return &noopLogger{} }

func (n *noopLogger) Debug(_ ...any)                          {}
func (n *noopLogger) Debugf(_ string, _ ...any)               {}
func (n *noopLogger) Debugw(_ string, _ ...any)               {}
func (n *noopLogger) Info(_ ...any)                           {}
func (n *noopLogger) Infof(_ string, _ ...any)                {}
func (n *noopLogger) Infow(_ string, _ ...any)                {}
func (n *noopLogger) Warn(_ ...any)                           {}
func (n *noopLogger) Warnf(_ string, _ ...any)                {}
func (n *noopLogger) Warnw(_ string, _ ...any)                {}
func (n *noopLogger) Error(_ ...any)                          {}
func (n *noopLogger) Errorf(_ string, _ ...any)               {}
func (n *noopLogger) Errorw(_ string, _ ...any)               {}
func (n *noopLogger) Fatal(_ ...any)                          {}
func (n *noopLogger) Fatalf(_ string, _ ...any)               {}
func (n *noopLogger) Fatalw(_ string, _ ...any)               {}
func (n *noopLogger) Debugc(_ context.Context, _ string, _ ...any) {}
func (n *noopLogger) Infoc(_ context.Context, _ string, _ ...any)  {}
func (n *noopLogger) Warnc(_ context.Context, _ string, _ ...any)  {}
func (n *noopLogger) Errorc(_ context.Context, _ string, _ ...any) {}
func (n *noopLogger) Fatalc(_ context.Context, _ string, _ ...any) {}
