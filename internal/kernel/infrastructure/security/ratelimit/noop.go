package ratelimit

import "context"

// NoopLimiter is a rate limiter that permits all requests.
// Useful in tests or environments where rate limiting is disabled.
type NoopLimiter struct{}

func (NoopLimiter) CheckIP(context.Context, string) error        { return nil }
func (NoopLimiter) CheckUser(context.Context, string) error      { return nil }
func (NoopLimiter) RecordFailedIP(context.Context, string) error   { return nil }
func (NoopLimiter) RecordFailedUser(context.Context, string) error { return nil }
func (NoopLimiter) ResetUser(context.Context, string) error        { return nil }
