package postgres

import (
	"context"

	"gct/pkg/logger"
	"go.uber.org/zap"
)

// Listen starts listening for notifications on a specific channel.
// onNotify is a callback function that handles the notification payload.
// It keeps running until the context is canceled or a critical error occurs.
func (p *Postgres) Listen(ctx context.Context, channel string, onNotify func(ctx context.Context, payload string) error, l logger.Log) {
	if p.Pool == nil {
		l.Warn("postgres connection is nil, skipping listener")
		return
	}

	// We need a dedicated connection for listening, acquired from the pool
	conn, err := p.Pool.Acquire(ctx)
	if err != nil {
		l.WithContext(ctx).Errorw("failed to acquire connection for listener", zap.Error(err))
		return
	}
	defer conn.Release()

	// Execute LISTEN command
	// Sanitize channel name if necessary, but here we assume strict internal usage
	_, err = conn.Exec(ctx, "LISTEN "+channel)
	if err != nil {
		l.WithContext(ctx).Errorw("failed to listen on channel", "channel", channel, zap.Error(err))
		return
	}

	l.WithContext(ctx).Infow("started listening for notifications", "channel", channel)

	for {
		select {
		case <-ctx.Done():
			l.WithContext(ctx).Infow("stopping listener", "channel", channel)
			return
		default:
			// WaitForNotification waits for a notification.
			// It will return an error if the context is canceled.
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				// check if it's just context cancellation
				if ctx.Err() != nil {
					l.WithContext(ctx).Errorw("error waiting for notification", zap.Error(err))
					return
				}
				// Connection might be broken, exit the loop.
				// In a real-world scenario, we might want a retry mechanism with backoff.
				return
			}

			l.WithContext(ctx).Infow("received notification", "channel", notification.Channel, "payload", notification.Payload)

			if notification.Payload != "" {
				if err := onNotify(ctx, notification.Payload); err != nil {
					l.WithContext(ctx).Errorw("failed to handle notification", "payload", notification.Payload, zap.Error(err))
				}
			}
		}
	}
}
