package audit

import "context"

// NoopLogger is a no-op implementation of Logger for tests and environments
// where audit logging is disabled.
type NoopLogger struct{}

// Log does nothing.
func (NoopLogger) Log(context.Context, Entry) {}
