package bitmap

import (
	"context"
)

// BitmapBasicI defines basic Redis Bitmap operations
type BitmapBasicI interface {
	SetBit(ctx context.Context, key string, offset int64, value int) (int64, error)
	GetBit(ctx context.Context, key string, offset int64) (int64, error)
	BitCount(ctx context.Context, key string, start, end int64) (int64, error)
	BitCountAll(ctx context.Context, key string) (int64, error)
	BitPos(ctx context.Context, key string, bit, start, end int64) (int64, error)
	BitPosAll(ctx context.Context, key string, bit int64) (int64, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// BitmapAdvancedI defines advanced Redis Bitmap operations
type BitmapAdvancedI interface {
	BitOpAnd(ctx context.Context, destKey string, keys ...string) (int64, error)
	BitOpOr(ctx context.Context, destKey string, keys ...string) (int64, error)
	BitOpXor(ctx context.Context, destKey string, keys ...string) (int64, error)
	BitOpNot(ctx context.Context, destKey, key string) (int64, error)
	BitField(ctx context.Context, key string, args ...any) ([]int64, error)
}

// BitmapI defines full Redis Bitmap operations interface
type BitmapI interface {
	BitmapBasicI
	BitmapAdvancedI
}
