package stream

import (
	"context"
	"fmt"
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
	result, err := s.client.XAdd(ctx, args).Result()
	if err != nil {
		return "", fmt.Errorf("failed to add message to stream %s: %w", stream, err)
	}
	return result, nil
}

// XAddWithID adds a new message to a stream with a specific ID
func (s *Stream) XAddWithID(ctx context.Context, stream, id string, values map[string]any) (string, error) {
	args := &redis.XAddArgs{
		Stream: stream,
		ID:     id,
		Values: values,
	}
	result, err := s.client.XAdd(ctx, args).Result()
	if err != nil {
		return "", fmt.Errorf("failed to add message with ID %s to stream %s: %w", id, stream, err)
	}
	return result, nil
}

// XAddWithMaxLen adds a message and maintains max length
func (s *Stream) XAddWithMaxLen(ctx context.Context, stream string, maxLen int64, values map[string]any) (string, error) {
	args := &redis.XAddArgs{
		Stream: stream,
		MaxLen: maxLen,
		Approx: true,
		Values: values,
	}
	result, err := s.client.XAdd(ctx, args).Result()
	if err != nil {
		return "", fmt.Errorf("failed to add message with max length to stream %s: %w", stream, err)
	}
	return result, nil
}

// XRead reads messages from streams
func (s *Stream) XRead(ctx context.Context, streams map[string]string) ([]redis.XStream, error) {
	args := &redis.XReadArgs{
		Streams: buildStreamsSlice(streams),
		Block:   0,
	}
	result, err := s.client.XRead(ctx, args).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read from streams %v: %w", streams, err)
	}
	return result, nil
}

// XReadWithBlock reads messages with blocking
func (s *Stream) XReadWithBlock(ctx context.Context, streams map[string]string, block time.Duration) ([]redis.XStream, error) {
	args := &redis.XReadArgs{
		Streams: buildStreamsSlice(streams),
		Block:   block,
	}
	result, err := s.client.XRead(ctx, args).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read streams with block: %w", err)
	}
	return result, nil
}

// XReadGroup reads messages from a consumer group
func (s *Stream) XReadGroup(ctx context.Context, group, consumer string, streams map[string]string) ([]redis.XStream, error) {
	args := &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  buildStreamsSlice(streams),
		Block:    0,
	}
	result, err := s.client.XReadGroup(ctx, args).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read from group %s: %w", group, err)
	}
	return result, nil
}

// XReadGroupWithBlock reads messages from a consumer group with blocking
func (s *Stream) XReadGroupWithBlock(ctx context.Context, group, consumer string, streams map[string]string, block time.Duration) ([]redis.XStream, error) {
	args := &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  buildStreamsSlice(streams),
		Block:    block,
	}
	result, err := s.client.XReadGroup(ctx, args).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read from group %s: %w", group, err)
	}
	return result, nil
}

// XGroupCreate creates a consumer group
func (s *Stream) XGroupCreate(ctx context.Context, stream, group, start string) error {
	if err := s.client.XGroupCreate(ctx, stream, group, start).Err(); err != nil {
		return fmt.Errorf("failed to create group %s in stream %s: %w", group, stream, err)
	}
	return nil
}

// XGroupCreateMkStream creates a consumer group and stream if it doesn't exist
func (s *Stream) XGroupCreateMkStream(ctx context.Context, stream, group, start string) error {
	if err := s.client.XGroupCreateMkStream(ctx, stream, group, start).Err(); err != nil {
		return fmt.Errorf("failed to create group %s with stream %s: %w", group, stream, err)
	}
	return nil
}

// XGroupDestroy destroys a consumer group
func (s *Stream) XGroupDestroy(ctx context.Context, stream, group string) error {
	if err := s.client.XGroupDestroy(ctx, stream, group).Err(); err != nil {
		return fmt.Errorf("failed to destroy group %s in stream %s: %w", group, stream, err)
	}
	return nil
}

// XAck acknowledges messages
func (s *Stream) XAck(ctx context.Context, stream, group string, ids ...string) (int64, error) {
	result, err := s.client.XAck(ctx, stream, group, ids...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to acknowledge messages in group %s: %w", group, err)
	}
	return result, nil
}

// XPending gets pending messages info
func (s *Stream) XPending(ctx context.Context, stream, group string) (*redis.XPending, error) {
	result, err := s.client.XPending(ctx, stream, group).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get pending messages for group %s: %w", group, err)
	}
	return result, nil
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
	result, err := s.client.XPendingExt(ctx, args).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get extended pending messages for group %s: %w", group, err)
	}
	return result, nil
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
	result, err := s.client.XClaim(ctx, args).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to claim messages for consumer %s: %w", consumer, err)
	}
	return result, nil
}

// XDel deletes messages from a stream
func (s *Stream) XDel(ctx context.Context, stream string, ids ...string) (int64, error) {
	result, err := s.client.XDel(ctx, stream, ids...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to delete messages from stream %s: %w", stream, err)
	}
	return result, nil
}

// XLen returns the length of a stream
func (s *Stream) XLen(ctx context.Context, stream string) (int64, error) {
	result, err := s.client.XLen(ctx, stream).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get length of stream %s: %w", stream, err)
	}
	return result, nil
}

// XRange reads a range of messages
func (s *Stream) XRange(ctx context.Context, stream, start, stop string) ([]redis.XMessage, error) {
	result, err := s.client.XRange(ctx, stream, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read range from stream %s: %w", stream, err)
	}
	return result, nil
}

// XRangeN reads N messages from a range
func (s *Stream) XRangeN(ctx context.Context, stream, start, stop string, count int64) ([]redis.XMessage, error) {
	result, err := s.client.XRangeN(ctx, stream, start, stop, count).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read range %d from stream %s: %w", count, stream, err)
	}
	return result, nil
}

// XRevRange reads a range of messages in reverse order
func (s *Stream) XRevRange(ctx context.Context, stream, start, stop string) ([]redis.XMessage, error) {
	result, err := s.client.XRevRange(ctx, stream, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read reverse range from stream %s: %w", stream, err)
	}
	return result, nil
}

// XRevRangeN reads N messages from a range in reverse order
func (s *Stream) XRevRangeN(ctx context.Context, stream, start, stop string, count int64) ([]redis.XMessage, error) {
	result, err := s.client.XRevRangeN(ctx, stream, start, stop, count).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to read reverse range %d from stream %s: %w", count, stream, err)
	}
	return result, nil
}

// XTrim trims the stream to a maximum length
func (s *Stream) XTrim(ctx context.Context, stream string, maxLen int64) (int64, error) {
	result, err := s.client.XTrimMaxLen(ctx, stream, maxLen).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to trim stream %s to max length %d: %w", stream, maxLen, err)
	}
	return result, nil
}

// XTrimApprox trims the stream approximately
func (s *Stream) XTrimApprox(ctx context.Context, stream string, maxLen int64) (int64, error) {
	result, err := s.client.XTrimMaxLenApprox(ctx, stream, maxLen, 0).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to trim stream %s approximately to max length %d: %w", stream, maxLen, err)
	}
	return result, nil
}

// buildStreamsSlice converts map to slice for XRead operations
func buildStreamsSlice(streams map[string]string) []string {
	result := make([]string, 0, len(streams)*2)
	for stream, id := range streams {
		result = append(result, stream, id)
	}
	return result
}
