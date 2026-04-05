package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"gct/internal/platform/application"
	"gct/internal/platform/domain"
	"gct/internal/platform/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

var _ application.EventBus = (*RedisStreamsEventBus)(nil)

// RedisStreamsEventBus implements EventBus using Redis Streams for persistence
// and Redis Pub/Sub for instant signaling.
type RedisStreamsEventBus struct {
	client    *redis.Client
	maxLen    int64
	log       logger.Log
	redisDown atomic.Bool
	mu        sync.RWMutex
	handlers  map[string][]application.EventHandler
}

// NewRedisStreamsEventBus creates a new Redis Streams-backed event bus.
func NewRedisStreamsEventBus(client *redis.Client, maxLen int64, log logger.Log) *RedisStreamsEventBus {
	return &RedisStreamsEventBus{
		client:   client,
		maxLen:   maxLen,
		log:      log,
		handlers: make(map[string][]application.EventHandler),
	}
}

// streamKey returns the Redis Stream key for an event name.
func streamKey(eventName string) string {
	return "stream:" + eventName
}

// signalChannel returns the Redis Pub/Sub channel for an event name.
func signalChannel(eventName string) string {
	return "signal:" + eventName
}

// eventPayload is the JSON structure stored in each stream entry.
type eventPayload struct {
	EventName   string    `json:"event_name"`
	AggregateID string    `json:"aggregate_id"`
	OccurredAt  time.Time `json:"occurred_at"`
}

// Publish writes each event to its Redis Stream and sends a Pub/Sub signal.
// Local handlers are also called synchronously (same as InMemoryEventBus).
func (b *RedisStreamsEventBus) Publish(ctx context.Context, events ...domain.DomainEvent) error {
	for _, event := range events {
		payload := eventPayload{
			EventName:   event.EventName(),
			AggregateID: event.AggregateID().String(),
			OccurredAt:  event.OccurredAt(),
		}

		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("marshal event %s: %w", event.EventName(), err)
		}

		// 1. XADD to stream (persistent)
		_, err = b.client.XAdd(ctx, &redis.XAddArgs{
			Stream: streamKey(event.EventName()),
			MaxLen: b.maxLen,
			Approx: true,
			Values: map[string]any{"data": string(data)},
		}).Result()
		if err != nil {
			if !b.redisDown.Load() {
				b.redisDown.Store(true)
				b.log.Warnw("Redis down, falling back to local handlers only",
					"event", event.EventName(), "error", err)
			}
		} else {
			if b.redisDown.Load() {
				b.redisDown.Store(false)
				b.log.Infow("Redis recovered, resuming stream publishing",
					"event", event.EventName())
			}
			// 2. PUBLISH signal (instant notification to subscribers)
			b.client.Publish(ctx, signalChannel(event.EventName()), "new")
		}

		// 3. Call local handlers (backward-compatible with InMemoryEventBus)
		b.mu.RLock()
		handlers := b.handlers[event.EventName()]
		b.mu.RUnlock()

		for _, handler := range handlers {
			if err := handler(ctx, event); err != nil {
				return err
			}
		}
	}
	return nil
}

// Subscribe registers a local handler for the given event name.
func (b *RedisStreamsEventBus) Subscribe(eventName string, handler application.EventHandler) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], handler)
	return nil
}

// ReadStream reads messages from a stream starting after lastID.
// Used by SSE handlers to fetch new messages after receiving a Pub/Sub signal.
func (b *RedisStreamsEventBus) ReadStream(ctx context.Context, stream string, lastID string, count int64) ([]redis.XMessage, error) {
	if lastID == "" {
		lastID = "0"
	}
	msgs, err := b.client.XRead(ctx, &redis.XReadArgs{
		Streams: []string{stream, lastID},
		Count:   count,
	}).Result()
	if err != nil {
		return nil, err
	}
	if len(msgs) == 0 {
		return nil, nil
	}
	return msgs[0].Messages, nil
}
