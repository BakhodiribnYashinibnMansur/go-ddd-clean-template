package dashboard

import (
	"context"
	"fmt"

	"gct/internal/domain"
)

// Get runs 7 aggregate COUNT queries and returns a DashboardStats struct.
func (r *Repo) Get(ctx context.Context) (domain.DashboardStats, error) {
	var stats domain.DashboardStats

	queries := []struct {
		dest *int64
		sql  string
	}{
		{&stats.TotalUsers, `SELECT COUNT(*) FROM users WHERE deleted_at = 0`},
		{&stats.ActiveSessions, `SELECT COUNT(*) FROM session WHERE expires_at > NOW()`},
		{&stats.AuditLogsToday, `SELECT COUNT(*) FROM audit_log WHERE created_at >= NOW()::date`},
		{&stats.SystemErrorsCount, `SELECT COUNT(*) FROM system_errors WHERE is_resolved = false`},
		{&stats.TotalFeatureFlags, `SELECT COUNT(*) FROM feature_flags WHERE deleted_at IS NULL AND is_active = true`},
		{&stats.TotalWebhooks, `SELECT COUNT(*) FROM webhooks WHERE deleted_at IS NULL AND is_active = true`},
		{&stats.TotalJobs, `SELECT COUNT(*) FROM jobs WHERE is_active = true`},
	}

	for _, q := range queries {
		if err := r.pool.QueryRow(ctx, q.sql).Scan(q.dest); err != nil {
			return domain.DashboardStats{}, fmt.Errorf("dashboard repo Get: %w", err)
		}
	}

	return stats, nil
}
