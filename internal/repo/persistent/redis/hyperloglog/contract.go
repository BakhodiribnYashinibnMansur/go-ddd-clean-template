package hyperloglog

import (
	"context"
)

// HyperLogLogI defines Redis HyperLogLog operations interface
type HyperLogLogI interface {
	PFAdd(ctx context.Context, key string, els ...any) (int64, error)
	PFCount(ctx context.Context, keys ...string) (int64, error)
	PFMerge(ctx context.Context, dest string, keys ...string) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}
