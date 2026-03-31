package sse

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// Bridge connects Redis Streams to the SSE Hub.
// It subscribes to a Pub/Sub signal channel and, on each signal,
// reads new messages from the corresponding Redis Stream and broadcasts them to the Hub.
type Bridge struct {
	client *redis.Client
	hub    *Hub
}

// NewBridge creates a new Bridge.
func NewBridge(client *redis.Client, hub *Hub) *Bridge {
	return &Bridge{client: client, hub: hub}
}

// Listen subscribes to the signalChannel and forwards new stream messages
// to the Hub under the given hubChannel. Blocks until ctx is cancelled.
func (b *Bridge) Listen(ctx context.Context, hubChannel, signalChannel, streamKey string) {
	ps := b.client.Subscribe(ctx, signalChannel)
	defer ps.Close()

	lastID := "0"
	psCh := ps.Channel()

	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-psCh:
			if !ok {
				return
			}
			// Read all new messages from the stream
			msgs, err := b.client.XRead(ctx, &redis.XReadArgs{
				Streams: []string{streamKey, lastID},
				Count:   100,
			}).Result()
			if err != nil {
				continue
			}
			for _, stream := range msgs {
				for _, msg := range stream.Messages {
					lastID = msg.ID
					data, ok := msg.Values["data"]
					if !ok {
						continue
					}
					b.hub.Broadcast(hubChannel, Message{
						ID:    msg.ID,
						Event: hubChannel,
						Data:  []byte(data.(string)),
					})
				}
			}
		}
	}
}
