package stream

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// StreamI defines Redis Stream operations interface
type StreamI interface {
	XAdd(ctx context.Context, stream string, values map[string]any) (string, error)
	XAddWithID(ctx context.Context, stream, id string, values map[string]any) (string, error)
	XAddWithMaxLen(ctx context.Context, stream string, maxLen int64, values map[string]any) (string, error)
	XRead(ctx context.Context, streams map[string]string) ([]redis.XStream, error)
	XReadWithBlock(ctx context.Context, streams map[string]string, block time.Duration) ([]redis.XStream, error)
	XReadGroup(ctx context.Context, group, consumer string, streams map[string]string) ([]redis.XStream, error)
	XReadGroupWithBlock(ctx context.Context, group, consumer string, streams map[string]string, block time.Duration) ([]redis.XStream, error)
	XGroupCreate(ctx context.Context, stream, group, start string) error
	XGroupCreateMkStream(ctx context.Context, stream, group, start string) error
	XGroupDestroy(ctx context.Context, stream, group string) error
	XAck(ctx context.Context, stream, group string, ids ...string) (int64, error)
	XPending(ctx context.Context, stream, group string) (*redis.XPending, error)
	XPendingExt(ctx context.Context, stream, group, start, end string, count int64) ([]redis.XPendingExt, error)
	XClaim(ctx context.Context, stream, group, consumer string, minIdleTime time.Duration, ids ...string) ([]redis.XMessage, error)
	XDel(ctx context.Context, stream string, ids ...string) (int64, error)
	XLen(ctx context.Context, stream string) (int64, error)
	XRange(ctx context.Context, stream, start, stop string) ([]redis.XMessage, error)
	XRangeN(ctx context.Context, stream, start, stop string, count int64) ([]redis.XMessage, error)
	XRevRange(ctx context.Context, stream, start, stop string) ([]redis.XMessage, error)
	XRevRangeN(ctx context.Context, stream, start, stop string, count int64) ([]redis.XMessage, error)
	XTrim(ctx context.Context, stream string, maxLen int64) (int64, error)
	XTrimApprox(ctx context.Context, stream string, maxLen int64) (int64, error)
}
