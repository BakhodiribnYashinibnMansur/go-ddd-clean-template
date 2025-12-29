package session

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var errNoSessionsToUpdate = errors.New("no sessions to update")

// SyncSessionActivityToPostgres syncs session last_activity from Redis to PostgreSQL
func (c *SessionCronJobs) SyncSessionActivityToPostgres() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	sessions, err := c.collectSessionActivities(ctx)
	if err != nil || len(sessions) == 0 {
		return
	}

	if err := c.syncToPostgres(ctx, sessions); err != nil {
		c.logger.Error("Failed to sync session activity to Postgres", zap.Error(err))
	}
}

func (c *SessionCronJobs) collectSessionActivities(ctx context.Context) ([]SessionActivity, error) {
	pattern := "session_last_activity:*"
	var keys []string
	iter := c.redis.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		c.logger.Errorw("Failed to scan Redis keys", "error", err, "pattern", pattern)
		return nil, err
	}

	if len(keys) == 0 {
		c.logger.Debug("No session activity keys found in Redis")
		return nil, nil
	}

	sessions := make([]SessionActivity, 0, len(keys))
	for _, key := range keys {
		activity, err := c.getSingleSessionActivity(ctx, key)
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				c.logger.Errorw("Failed to get session activity", "error", err, "key", key)
			}
			continue
		}
		sessions = append(sessions, *activity)
	}
	return sessions, nil
}

func (c *SessionCronJobs) getSingleSessionActivity(ctx context.Context, key string) (*SessionActivity, error) {
	lastActivityStr, err := c.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	lastActivity, err := time.Parse(time.RFC3339, lastActivityStr)
	if err != nil {
		return nil, err
	}

	return &SessionActivity{
		SessionID:    strings.TrimPrefix(key, "session_last_activity:"),
		LastActivity: lastActivity,
	}, nil
}

func (c *SessionCronJobs) syncToPostgres(ctx context.Context, sessions []SessionActivity) error {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := c.batchUpdateSessionActivity(ctx, tx, sessions); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	c.logger.Info("Session activity synced to PostgreSQL", zap.Int("count", len(sessions)))
	return nil
}

// batchUpdateSessionActivity performs batch update using temporary table
func (c *SessionCronJobs) batchUpdateSessionActivity(ctx context.Context, tx pgx.Tx, sessions []SessionActivity) error {
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
		c.logger.Error("Failed to create temp table", zap.Error(err))
		return err
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
		c.logger.Error("Failed to copy data to temp table", zap.Error(err))
		return err
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
		c.logger.Error("Failed to update sessions from temp table", zap.Error(err))
		return err
	}

	rowsAffected := result.RowsAffected()
	c.logger.Debug("Updated sessions from Redis",
		zap.Int64("rows_affected", rowsAffected),
	)

	return nil
}
