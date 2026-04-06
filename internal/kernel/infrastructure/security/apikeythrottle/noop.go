package apikeythrottle

import "context"

// NoopThrottle is a no-op implementation that never throttles. Useful for
// development or when Redis is unavailable.
type NoopThrottle struct{}

// Check always returns nil.
func (NoopThrottle) Check(_ context.Context, _ string) error { return nil }

// RecordFail always returns nil.
func (NoopThrottle) RecordFail(_ context.Context, _ string) error { return nil }
