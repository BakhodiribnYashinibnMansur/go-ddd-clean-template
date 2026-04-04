// Package logstore contains helpers shared by log-like tables that use monthly
// range partitioning (app_logs, http_request_logs, external_api_logs).
package logstore

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PartitionName returns the child partition name for `table` and month `t`
// (e.g. "http_request_logs_2026_04").
func PartitionName(table string, t time.Time) string {
	return fmt.Sprintf("%s_%s", table, t.Format("2006_01"))
}

// EnsureFuture creates partitions for the current month and the next `ahead`
// months if they do not already exist. Call periodically to keep the write
// path from hitting "no partition of relation" errors.
func EnsureFuture(ctx context.Context, pool *pgxpool.Pool, table string, ahead int) error {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i <= ahead; i++ {
		lo := start.AddDate(0, i, 0)
		hi := start.AddDate(0, i+1, 0)
		name := PartitionName(table, lo)
		stmt := fmt.Sprintf(
			`CREATE TABLE IF NOT EXISTS %s PARTITION OF %s FOR VALUES FROM ('%s') TO ('%s')`,
			name, table,
			lo.Format("2006-01-02"), hi.Format("2006-01-02"),
		)
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("logstore: create partition %s: %w", name, err)
		}
	}
	return nil
}

// DropOlderThan drops any monthly partitions of `table` whose suffix (YYYY_MM)
// is strictly older than the cutoff month. Uses DROP TABLE which is O(1)
// regardless of row count — the right way to expire time-series data.
//
// Returns the number of partitions dropped.
func DropOlderThan(ctx context.Context, pool *pgxpool.Pool, table string, cutoff time.Time) (int, error) {
	// Find partitions belonging to `table` that are older than cutoff. We
	// compare suffixes lexicographically because YYYY_MM sorts correctly.
	cutoffSuffix := cutoff.Format("2006_01")
	prefix := table + "_"

	q := `
		SELECT child.relname
		FROM pg_inherits i
		JOIN pg_class parent ON parent.oid = i.inhparent
		JOIN pg_class child  ON child.oid  = i.inhrelid
		WHERE parent.relname = $1
		  AND child.relname LIKE $2
		  AND substring(child.relname FROM length($1) + 2) < $3
	`
	rows, err := pool.Query(ctx, q, table, prefix+"%", cutoffSuffix)
	if err != nil {
		return 0, fmt.Errorf("logstore: list partitions: %w", err)
	}
	defer rows.Close()

	var victims []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return 0, err
		}
		victims = append(victims, name)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	for _, v := range victims {
		if _, err := pool.Exec(ctx, fmt.Sprintf(`DROP TABLE IF EXISTS %s`, v)); err != nil {
			return 0, fmt.Errorf("logstore: drop %s: %w", v, err)
		}
	}
	return len(victims), nil
}
