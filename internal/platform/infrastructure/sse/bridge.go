package sse

import (
	"context"
	"time"

	"gct/internal/platform/infrastructure/logger"

	"github.com/redis/go-redis/v9"
)

// Bridge connects Redis Streams to the SSE Hub.
// It subscribes to a Pub/Sub signal channel and, on each signal,
// reads new messages from the corresponding Redis Stream and broadcasts them to the Hub.
type Bridge struct {
	client *redis.Client
	hub    *Hub
	log    logger.Log
}

// NewBridge creates a new Bridge.
func NewBridge(client *redis.Client, hub *Hub, log logger.Log) *Bridge {
	return &Bridge{client: client, hub: hub, log: log}
}

// Listen subscribes to the signalChannel and forwards new stream messages
// to the Hub under the given hubChannel. Blocks until ctx is cancelled.
// It automatically reconnects with exponential backoff if the connection drops.
func (b *Bridge) Listen(ctx context.Context, hubChannel, signalChannel, streamKey string) {
	const (
		initialBackoff = 1 * time.Second
		maxBackoff     = 30 * time.Second
	)
	backoff := initialBackoff

	for {
		select {
		case <-ctx.Done():
			b.log.Infow("SSE Bridge shutting down", "channel", hubChannel)
			return
		default:
		}

		ps := b.client.Subscribe(ctx, signalChannel)

		// Reset backoff on successful subscribe
		b.log.Infow("SSE Bridge connected", "channel", signalChannel)
		backoff = initialBackoff

		lastID := "0"
		psCh := ps.Channel()

		reconnect := false
		for !reconnect {
			select {
			case <-ctx.Done():
				ps.Close()
				return
			case _, ok := <-psCh:
				if !ok {
					b.log.Warnw("SSE Bridge channel closed, reconnecting",
						"channel", signalChannel)
					reconnect = true
					break
				}
				msgs, err := b.client.XRead(ctx, &redis.XReadArgs{
					Streams: []string{streamKey, lastID},
					Count:   100,
				}).Result()
				if err != nil {
					b.log.Warnw("SSE Bridge XRead error",
						"stream", streamKey, "error", err)
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
		ps.Close()

		// Backoff before reconnect
		b.log.Warnw("SSE Bridge reconnecting with backoff",
			"channel", signalChannel, "backoff", backoff)
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, maxBackoff)
	}
}
