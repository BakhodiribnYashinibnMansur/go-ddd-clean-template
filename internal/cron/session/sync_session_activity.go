package session

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

var errNoSessionsToUpdate = errors.New("no sessions to update")

// SyncSessionActivityToPostgres syncs session last_activity from Redis to PostgreSQL
func (c *CronJobs) SyncSessionActivityToPostgres(ctx context.Context) {
	sessions, err := c.collectSessionActivities(ctx)
	if err != nil || len(sessions) == 0 {
		return
	}

	if err := c.syncToPostgres(ctx, sessions); err != nil {
		c.logger.Errorc(ctx, "Failed to sync session activity to Postgres", "error", err)
	}
}

func (c *CronJobs) collectSessionActivities(ctx context.Context) ([]Activity, error) {
	pattern := "session_last_activity:*"
	var keys []string
	iter := c.redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		c.logger.Errorc(ctx, "Failed to scan Redis keys", "error", err, "pattern", pattern)
		return nil, fmt.Errorf("redis scan error: %w", err)
	}

	if len(keys) == 0 {
		c.logger.Debugc(ctx, "No session activity keys found in Redis")
		return nil, nil
	}

	sessions := make([]Activity, 0, len(keys))
	for _, key := range keys {
		activity, err := c.getSingleSessionActivity(ctx, key)
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				c.logger.Errorc(ctx, "Failed to get session activity", "error", err, "key", key)
			}
			continue
		}
		sessions = append(sessions, *activity)
	}
	return sessions, nil
}

func (c *CronJobs) getSingleSessionActivity(ctx context.Context, key string) (*Activity, error) {
	lastActivityStr, err := c.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get from redis: %w", err)
	}

	lastActivity, err := time.Parse(time.RFC3339, lastActivityStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse activity time: %w", err)
	}

	return &Activity{
		SessionID:    strings.TrimPrefix(key, "session_last_activity:"),
		LastActivity: lastActivity,
	}, nil
}

func (c *CronJobs) syncToPostgres(ctx context.Context, sessions []Activity) error {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := c.batchUpdateSessionActivity(ctx, tx, sessions); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	c.logger.Infoc(ctx, "Session activity synced to PostgreSQL", "count", len(sessions))
	return nil
}

// batchUpdateSessionActivity performs batch update using temporary table
func (c *CronJobs) batchUpdateSessionActivity(ctx context.Context, tx pgx.Tx, sessions []Activity) error {
	if len(sessions) == 0 {
		return errNoSessionsToUpdate
	}

	// Create temporary table
	_, err := tx.Exec(ctx, `
		CREATE TEMP TABLE temp_session_activity (
			session_id UUID,
			last_activity TIMESTAMP
		) ON COMMIT DROP
	`)
	if err != nil {
		c.logger.Errorc(ctx, "Failed to create temp table", "error", err)
		return fmt.Errorf("failed to create temp table: %w", err)
	}

	// Prepare data for CopyFrom
	rows := make([][]any, 0, len(sessions))
	for _, session := range sessions {
		rows = append(rows, []any{session.SessionID, session.LastActivity})
	}

	// Batch insert using CopyFrom
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"temp_session_activity"},
		[]string{"session_id", "last_activity"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		c.logger.Errorc(ctx, "Failed to copy data to temp table", "error", err)
		return fmt.Errorf("failed to copy data to temp table: %w", err)
	}

	// Update main table from temp table
	result, err := tx.Exec(ctx, `
		UPDATE session AS s
		SET 
			last_activity = t.last_activity,
			updated_at = NOW()
		FROM temp_session_activity AS t
		WHERE s.id = t.session_id::UUID
	`)
	if err != nil {
		c.logger.Errorc(ctx, "Failed to update sessions from temp table", "error", err)
		return fmt.Errorf("failed to update sessions from temp table: %w", err)
	}

	rowsAffected := result.RowsAffected()
	c.logger.Debugc(ctx, "Updated sessions from Redis",
		"rows_affected", rowsAffected,
	)

	return nil
}
