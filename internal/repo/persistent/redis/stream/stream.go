package stream

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Stream handles Redis Stream operations
type Stream struct {
	client *redis.Client
}

// New creates a new Stream instance
func New(client *redis.Client) *Stream {
	return &Stream{
		client: client,
	}
}

// XAdd adds a new message to a stream
func (s *Stream) XAdd(ctx context.Context, stream string, values map[string]any) (string, error) {
	args := &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}
	return s.client.XAdd(ctx, args).Result()
}

// XAddWithID adds a new message to a stream with a specific ID
func (s *Stream) XAddWithID(ctx context.Context, stream, id string, values map[string]any) (string, error) {
	args := &redis.XAddArgs{
		Stream: stream,
		ID:     id,
		Values: values,
	}
	return s.client.XAdd(ctx, args).Result()
}

// XAddWithMaxLen adds a message and maintains max length
func (s *Stream) XAddWithMaxLen(ctx context.Context, stream string, maxLen int64, values map[string]any) (string, error) {
	args := &redis.XAddArgs{
		Stream: stream,
		MaxLen: maxLen,
		Approx: true,
		Values: values,
	}
	return s.client.XAdd(ctx, args).Result()
}

// XRead reads messages from streams
func (s *Stream) XRead(ctx context.Context, streams map[string]string) ([]redis.XStream, error) {
	args := &redis.XReadArgs{
		Streams: buildStreamsSlice(streams),
		Block:   0,
	}
	return s.client.XRead(ctx, args).Result()
}

// XReadWithBlock reads messages with blocking
func (s *Stream) XReadWithBlock(ctx context.Context, streams map[string]string, block time.Duration) ([]redis.XStream, error) {
	args := &redis.XReadArgs{
		Streams: buildStreamsSlice(streams),
		Block:   block,
	}
	return s.client.XRead(ctx, args).Result()
}

// XReadGroup reads messages from a consumer group
func (s *Stream) XReadGroup(ctx context.Context, group, consumer string, streams map[string]string) ([]redis.XStream, error) {
	args := &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  buildStreamsSlice(streams),
		Block:    0,
	}
	return s.client.XReadGroup(ctx, args).Result()
}

// XReadGroupWithBlock reads messages from a consumer group with blocking
func (s *Stream) XReadGroupWithBlock(ctx context.Context, group, consumer string, streams map[string]string, block time.Duration) ([]redis.XStream, error) {
	args := &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  buildStreamsSlice(streams),
		Block:    block,
	}
	return s.client.XReadGroup(ctx, args).Result()
}

// XGroupCreate creates a consumer group
func (s *Stream) XGroupCreate(ctx context.Context, stream, group, start string) error {
	return s.client.XGroupCreate(ctx, stream, group, start).Err()
}

// XGroupCreateMkStream creates a consumer group and stream if it doesn't exist
func (s *Stream) XGroupCreateMkStream(ctx context.Context, stream, group, start string) error {
	return s.client.XGroupCreateMkStream(ctx, stream, group, start).Err()
}

// XGroupDestroy destroys a consumer group
func (s *Stream) XGroupDestroy(ctx context.Context, stream, group string) error {
	return s.client.XGroupDestroy(ctx, stream, group).Err()
}

// XAck acknowledges messages
func (s *Stream) XAck(ctx context.Context, stream, group string, ids ...string) (int64, error) {
	return s.client.XAck(ctx, stream, group, ids...).Result()
}

// XPending gets pending messages info
func (s *Stream) XPending(ctx context.Context, stream, group string) (*redis.XPending, error) {
	return s.client.XPending(ctx, stream, group).Result()
}

// XPendingExt gets extended pending messages info
func (s *Stream) XPendingExt(ctx context.Context, stream, group, start, end string, count int64) ([]redis.XPendingExt, error) {
	args := &redis.XPendingExtArgs{
		Stream: stream,
		Group:  group,
		Start:  start,
		End:    end,
		Count:  count,
	}
	return s.client.XPendingExt(ctx, args).Result()
}

// XClaim claims pending messages
func (s *Stream) XClaim(ctx context.Context, stream, group, consumer string, minIdleTime time.Duration, ids ...string) ([]redis.XMessage, error) {
	args := &redis.XClaimArgs{
		Stream:   stream,
		Group:    group,
		Consumer: consumer,
		MinIdle:  minIdleTime,
		Messages: ids,
	}
	return s.client.XClaim(ctx, args).Result()
}

// XDel deletes messages from a stream
func (s *Stream) XDel(ctx context.Context, stream string, ids ...string) (int64, error) {
	return s.client.XDel(ctx, stream, ids...).Result()
}

// XLen returns the length of a stream
func (s *Stream) XLen(ctx context.Context, stream string) (int64, error) {
	return s.client.XLen(ctx, stream).Result()
}

// XRange reads a range of messages
func (s *Stream) XRange(ctx context.Context, stream, start, stop string) ([]redis.XMessage, error) {
	return s.client.XRange(ctx, stream, start, stop).Result()
}

// XRangeN reads N messages from a range
func (s *Stream) XRangeN(ctx context.Context, stream, start, stop string, count int64) ([]redis.XMessage, error) {
	return s.client.XRangeN(ctx, stream, start, stop, count).Result()
}

// XRevRange reads a range of messages in reverse order
func (s *Stream) XRevRange(ctx context.Context, stream, start, stop string) ([]redis.XMessage, error) {
	return s.client.XRevRange(ctx, stream, start, stop).Result()
}

// XRevRangeN reads N messages from a range in reverse order
func (s *Stream) XRevRangeN(ctx context.Context, stream, start, stop string, count int64) ([]redis.XMessage, error) {
	return s.client.XRevRangeN(ctx, stream, start, stop, count).Result()
}

// XTrim trims the stream to a maximum length
func (s *Stream) XTrim(ctx context.Context, stream string, maxLen int64) (int64, error) {
	return s.client.XTrimMaxLen(ctx, stream, maxLen).Result()
}

// XTrimApprox trims the stream approximately
func (s *Stream) XTrimApprox(ctx context.Context, stream string, maxLen int64) (int64, error) {
	return s.client.XTrimMaxLenApprox(ctx, stream, maxLen, 0).Result()
}

// buildStreamsSlice converts map to slice for XRead operations
func buildStreamsSlice(streams map[string]string) []string {
	result := make([]string, 0, len(streams)*2)
	for stream, id := range streams {
		result = append(result, stream, id)
	}
	return result
}
