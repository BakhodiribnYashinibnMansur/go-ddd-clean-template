package session

import (
	"context"
	"time"
)

// ExpireOldSessions marks sessions older than 30 days as expired
func (c *CronJobs) ExpireOldSessions() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Calculate the expiration timestamp (30 days ago)
	expirationTime := time.Now().AddDate(0, 0, -SessionExpirationDays)

	c.logger.WithContext(ctx).Infow("Starting session expiration process",
		"expiration_time", expirationTime,
		"days", SessionExpirationDays,
	)

	// Revoke sessions that haven't had activity in the last 30 days
	query := `
		UPDATE sessions
		SET 
			revoked = true,
			updated_at = NOW()
		WHERE 
			revoked = false 
			AND last_activity < $1
	`

	result, err := c.pool.Exec(ctx, query, expirationTime)
	if err != nil {
		c.logger.WithContext(ctx).Errorw("Failed to expire old sessions",
			"error", err,
			"expiration_time", expirationTime,
		)
		return
	}

	rowsAffected := result.RowsAffected()

	c.logger.WithContext(ctx).Infow("Session expiration process completed",
		"expired_count", rowsAffected,
		"duration", time.Since(expirationTime),
	)

	// Log warning if too many sessions were expired
	if rowsAffected > 1000 {
		c.logger.WithContext(ctx).Warnw("Large number of sessions expired",
			"count", rowsAffected,
		)
	}
}
