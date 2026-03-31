package pubsub

import (
	"context"

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
// Blocks until ctx is cancelled.
func (s *Subscriber) Subscribe(ctx context.Context, channel string, handler MessageHandler) {
	ps := s.client.Subscribe(ctx, channel)
	defer ps.Close()

	ch := ps.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			handler(msg.Channel, msg.Payload)
		}
	}
}
