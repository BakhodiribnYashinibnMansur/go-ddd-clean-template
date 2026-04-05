package pubsub

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// MessageHandler is called when a message arrives on a subscribed channel.
type MessageHandler func(channel, payload string)

// Subscriber wraps Redis Pub/Sub for listening to signal channels.
type Subscriber struct {
	client *redis.Client
}

// NewSubscriber creates a new Pub/Sub subscriber.
func NewSubscriber(client *redis.Client) *Subscriber {
	return &Subscriber{client: client}
}

// Subscribe listens on the given channel and calls handler for each message.
// Blocks until ctx is cancelled. Reconnects automatically on connection loss.
func (s *Subscriber) Subscribe(ctx context.Context, channel string, handler MessageHandler) {
	s.listenWithRetry(ctx, func() *redis.PubSub {
		return s.client.Subscribe(ctx, channel)
	}, handler, channel)
}

// PSubscribe listens on channels matching the given glob pattern and calls
// handler for each message. Blocks until ctx is cancelled. Reconnects automatically on connection loss.
func (s *Subscriber) PSubscribe(ctx context.Context, pattern string, handler MessageHandler) {
	s.listenWithRetry(ctx, func() *redis.PubSub {
		return s.client.PSubscribe(ctx, pattern)
	}, handler, pattern)
}

// listenWithRetry handles subscription with automatic reconnection on failure.
func (s *Subscriber) listenWithRetry(ctx context.Context, subscribeFn func() *redis.PubSub, handler MessageHandler, name string) {
	backoff := time.Second
	const maxBackoff = 30 * time.Second

	for {
		if ctx.Err() != nil {
			return
		}

		ps := subscribeFn()

		// Verify the subscription succeeded
		if _, err := ps.Receive(ctx); err != nil {
			ps.Close()
			log.Printf("pubsub: failed to subscribe to %s: %v, retrying in %v", name, err, backoff)
			select {
			case <-ctx.Done():
				return
			case <-time.After(backoff):
				backoff = min(backoff*2, maxBackoff)
				continue
			}
		}
		backoff = time.Second

		ch := ps.Channel()
		func() {
			defer ps.Close()
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-ch:
					if !ok {
						log.Printf("pubsub: channel %s closed, reconnecting", name)
						return
					}
					handler(msg.Channel, msg.Payload)
				}
			}
		}()
	}
}
