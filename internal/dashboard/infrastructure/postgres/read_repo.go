package postgres

import (
	"context"

	appdto "gct/internal/dashboard/application"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DashboardReadRepo implements query.DashboardReadRepository for the CQRS read side.
type DashboardReadRepo struct {
	pool *pgxpool.Pool
}

// NewDashboardReadRepo creates a new DashboardReadRepo.
func NewDashboardReadRepo(pool *pgxpool.Pool) *DashboardReadRepo {
	return &DashboardReadRepo{pool: pool}
}

// GetStats returns aggregated dashboard statistics by running COUNT queries against multiple tables.
func (r *DashboardReadRepo) GetStats(ctx context.Context) (*appdto.DashboardStatsView, error) {
	view := &appdto.DashboardStatsView{}

	// Total users (soft-delete aware).
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableUsers+" WHERE deleted_at = 0",
	).Scan(&view.TotalUsers); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableUsers, nil)
	}

	// Active (non-revoked, non-expired) sessions.
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableSession+" WHERE revoked = false AND expires_at > NOW()",
	).Scan(&view.ActiveSessions); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableSession, nil)
	}

	// Audit log entries created today.
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableAuditLog+" WHERE created_at >= CURRENT_DATE",
	).Scan(&view.AuditLogsToday); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableAuditLog, nil)
	}

	// Unresolved system errors.
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableSystemError+" WHERE is_resolved = false",
	).Scan(&view.SystemErrorsCount); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableSystemError, nil)
	}

	// Total feature flags.
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableFeatureFlags,
	).Scan(&view.TotalFeatureFlags); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableFeatureFlags, nil)
	}

	// Total webhooks.
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableWebhooks,
	).Scan(&view.TotalWebhooks); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableWebhooks, nil)
	}

	// Total jobs.
	if err := r.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM "+consts.TableJobs,
	).Scan(&view.TotalJobs); err != nil {
		return nil, apperrors.HandlePgError(err, consts.TableJobs, nil)
	}

	return view, nil
}
