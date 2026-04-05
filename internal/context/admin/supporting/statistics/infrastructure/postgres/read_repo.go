package postgres

import (
	"context"

	appdto "gct/internal/context/admin/supporting/statistics/application"
	"gct/internal/kernel/consts"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StatisticsReadRepo implements query.StatisticsReadRepository for the CQRS read side.
type StatisticsReadRepo struct {
	pool *pgxpool.Pool
}

// NewStatisticsReadRepo creates a new StatisticsReadRepo.
func NewStatisticsReadRepo(pool *pgxpool.Pool) *StatisticsReadRepo {
	return &StatisticsReadRepo{pool: pool}
}

// GetOverview returns the top-level aggregated counts.
func (r *StatisticsReadRepo) GetOverview(ctx context.Context) (result *appdto.OverviewView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "StatisticsReadRepo.GetOverview")
	defer func() { end(err) }()

	view := &appdto.OverviewView{}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableUsers+" WHERE deleted_at = 0",
	).Scan(&view.TotalUsers); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableUsers, nil)
	}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableSession+" WHERE revoked = false AND expires_at > NOW()",
	).Scan(&view.ActiveSessions); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableSession, nil)
	}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableAuditLog+" WHERE created_at >= CURRENT_DATE",
	).Scan(&view.AuditLogsToday); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableAuditLog, nil)
	}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableSystemError+" WHERE is_resolved = false",
	).Scan(&view.SystemErrorsCount); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableSystemError, nil)
	}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableFeatureFlags,
	).Scan(&view.TotalFeatureFlags); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableFeatureFlags, nil)
	}

	return view, nil
}

// GetUserStats returns the user lifecycle and role breakdown.
func (r *StatisticsReadRepo) GetUserStats(ctx context.Context) (result *appdto.UserStatsView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "StatisticsReadRepo.GetUserStats")
	defer func() { end(err) }()

	view := &appdto.UserStatsView{ByRole: make(map[string]int64)}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableUsers+" WHERE deleted_at = 0",
	).Scan(&view.Total); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableUsers, nil)
	}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableUsers+" WHERE deleted_at <> 0",
	).Scan(&view.Deleted); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableUsers, nil)
	}

	rows, err := r.pool.Query(ctx,
		"SELECT r.name, COUNT(u.id) FROM "+consts.TableUsers+" u "+
			"JOIN "+consts.TableRole+" r ON u.role_id = r.id "+
			"WHERE u.deleted_at = 0 "+
			"GROUP BY r.name",
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableUsers, nil)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		var count int64
		if err := rows.Scan(&name, &count); err != nil {
			return nil, apperrors.HandlePgError(err, consts.TableUsers, nil)
		}
		view.ByRole[name] = count
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableUsers, nil)
	}

	return view, nil
}

// GetSessionStats returns the session state breakdown.
func (r *StatisticsReadRepo) GetSessionStats(ctx context.Context) (result *appdto.SessionStatsView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "StatisticsReadRepo.GetSessionStats")
	defer func() { end(err) }()

	view := &appdto.SessionStatsView{}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableSession+" WHERE revoked = false AND expires_at > NOW()",
	).Scan(&view.Active); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableSession, nil)
	}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableSession+" WHERE revoked = false AND expires_at <= NOW()",
	).Scan(&view.Expired); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableSession, nil)
	}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableSession+" WHERE revoked = true",
	).Scan(&view.Revoked); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableSession, nil)
	}

	return view, nil
}

// GetErrorStats returns the system error breakdown.
func (r *StatisticsReadRepo) GetErrorStats(ctx context.Context) (result *appdto.ErrorStatsView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "StatisticsReadRepo.GetErrorStats")
	defer func() { end(err) }()

	view := &appdto.ErrorStatsView{}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableSystemError+" WHERE is_resolved = false",
	).Scan(&view.Unresolved); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableSystemError, nil)
	}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableSystemError+" WHERE is_resolved = true",
	).Scan(&view.Resolved); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableSystemError, nil)
	}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableSystemError+" WHERE created_at >= NOW() - INTERVAL '24 hours'",
	).Scan(&view.Last24h); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableSystemError, nil)
	}

	return view, nil
}

// GetAuditStats returns the audit log recency breakdown.
func (r *StatisticsReadRepo) GetAuditStats(ctx context.Context) (result *appdto.AuditStatsView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "StatisticsReadRepo.GetAuditStats")
	defer func() { end(err) }()

	view := &appdto.AuditStatsView{}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableAuditLog+" WHERE created_at >= CURRENT_DATE",
	).Scan(&view.Today); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableAuditLog, nil)
	}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableAuditLog+" WHERE created_at >= CURRENT_DATE - INTERVAL '7 days'",
	).Scan(&view.Last7Days); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableAuditLog, nil)
	}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableAuditLog,
	).Scan(&view.Total); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableAuditLog, nil)
	}

	return view, nil
}

// GetSecurityStats returns counts for ip_rules and rate_limits.
func (r *StatisticsReadRepo) GetSecurityStats(ctx context.Context) (result *appdto.SecurityStatsView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "StatisticsReadRepo.GetSecurityStats")
	defer func() { end(err) }()

	view := &appdto.SecurityStatsView{}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableIPRules,
	).Scan(&view.IPRules); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableIPRules, nil)
	}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableRateLimits,
	).Scan(&view.RateLimits); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableRateLimits, nil)
	}

	return view, nil
}

// GetFeatureFlagStats returns the feature flag active-state breakdown.
// Note: feature_flags uses is_active as the enable flag.
func (r *StatisticsReadRepo) GetFeatureFlagStats(ctx context.Context) (result *appdto.FeatureFlagStatsView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "StatisticsReadRepo.GetFeatureFlagStats")
	defer func() { end(err) }()

	view := &appdto.FeatureFlagStatsView{}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableFeatureFlags,
	).Scan(&view.Total); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableFeatureFlags, nil)
	}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableFeatureFlags+" WHERE is_active = true",
	).Scan(&view.Enabled); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableFeatureFlags, nil)
	}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableFeatureFlags+" WHERE is_active = false",
	).Scan(&view.Disabled); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableFeatureFlags, nil)
	}

	return view, nil
}

// GetContentStats returns counts for content tables.
func (r *StatisticsReadRepo) GetContentStats(ctx context.Context) (result *appdto.ContentStatsView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "StatisticsReadRepo.GetContentStats")
	defer func() { end(err) }()

	view := &appdto.ContentStatsView{}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableAnnouncements,
	).Scan(&view.Announcements); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableAnnouncements, nil)
	}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableNotifications,
	).Scan(&view.Notifications); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableNotifications, nil)
	}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableFileMetadata,
	).Scan(&view.FileMetadata); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableFileMetadata, nil)
	}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableTranslations,
	).Scan(&view.Translations); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableTranslations, nil)
	}

	return view, nil
}

// GetIntegrationStats returns counts for integrations and api_keys (soft-delete aware).
func (r *StatisticsReadRepo) GetIntegrationStats(ctx context.Context) (result *appdto.IntegrationStatsView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "StatisticsReadRepo.GetIntegrationStats")
	defer func() { end(err) }()

	view := &appdto.IntegrationStatsView{}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableIntegrations+" WHERE deleted_at IS NULL",
	).Scan(&view.Integrations); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableIntegrations, nil)
	}

	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableAPIKeys+" WHERE deleted_at IS NULL",
	).Scan(&view.APIKeys); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableAPIKeys, nil)
	}

	return view, nil
}
