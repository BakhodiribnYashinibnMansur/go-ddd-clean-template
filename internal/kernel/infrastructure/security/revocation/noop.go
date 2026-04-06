package revocation

import (
	"context"
	"time"
)

// NoopStore is a no-op implementation useful for testing or when
// revocation is disabled.
type NoopStore struct{}

func (NoopStore) Revoke(context.Context, string, time.Duration) error   { return nil }
func (NoopStore) IsRevoked(context.Context, string) bool                { return false }
func (NoopStore) RevokeMany(context.Context, []string, time.Duration) error { return nil }
