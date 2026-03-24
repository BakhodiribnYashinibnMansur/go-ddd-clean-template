package postgres

import (
	"context"

	"gct/internal/shared/infrastructure/logger"
)

// Listen starts listening for notifications on a specific channel.
// onNotify is a callback function that handles the notification payload.
// It keeps running until the context is canceled or a critical error occurs.
func (p *Postgres) Listen(ctx context.Context, channel string, onNotify func(ctx context.Context, payload string) error, l logger.Log) {
	if p.Pool == nil {
		l.Warn("⚠️  PostgreSQL connection is nil, skipping listener")
		return
	}

	// We need a dedicated connection for listening, acquired from the pool
	conn, err := p.Pool.Acquire(ctx)
	if err != nil {
		l.Errorc(ctx, "❌ Failed to acquire connection for listener", "error", err, "channel", channel)
		return
	}
	defer conn.Release()

	// Execute LISTEN command
	// Sanitize channel name if necessary, but here we assume strict internal usage
	_, err = conn.Exec(ctx, "LISTEN "+channel)
	if err != nil {
		l.Errorc(ctx, "❌ Failed to execute LISTEN command", "channel", channel, "error", err)
		return
	}

	l.Infoc(ctx, "👂 Started listening for PostgreSQL notifications", "channel", channel)

	for {
		select {
		case <-ctx.Done():
			l.Infoc(ctx, "🛑 Stopping PostgreSQL listener", "channel", channel)
			return
		default:
			// WaitForNotification waits for a notification.
			// It will return an error if the context is canceled.
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				// check if it's just context cancellation
				if ctx.Err() != nil {
					l.Errorc(ctx, "⚠️  PostgreSQL listener context canceled", "error", err, "channel", channel)
					return
				}
				// Connection might be broken, exit the loop.
				// In a real-world scenario, we might want a retry mechanism with backoff.
				l.Errorc(ctx, "❌ PostgreSQL listener connection error", "error", err, "channel", channel)
				return
			}

			l.Infoc(ctx, "📬 Received PostgreSQL notification", "channel", notification.Channel, "payload", notification.Payload)

			if notification.Payload != "" {
				if err := onNotify(ctx, notification.Payload); err != nil {
					l.Errorc(ctx, "❌ Failed to handle notification", "payload", notification.Payload, "error", err)
				}
			}
		}
	}
}
