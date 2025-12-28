package session

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/evrone/go-clean-template/internal/domain"
	apperrors "github.com/evrone/go-clean-template/pkg/errors"
)

func (r *Repo) Update(ctx context.Context, s *domain.Session) error {
	sql, args, err := r.builder.
		Update("session").
		Set("fcm_token", s.FCMToken).
		Set("revoked", s.Revoked).
		Set("last_activity", s.LastActivity).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": s.ID}).
		ToSql()
	if err != nil {
		return apperrors.AutoSource(
			apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
				"failed to build update SQL query")).
			WithField("id", s.ID.String()).
			WithDetails("Error occurred while building UPDATE query for session")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(ctx, err, "session", map[string]any{
			"id": s.ID.String(),
		})
	}

	return nil
}

func (r *Repo) Revoke(ctx context.Context, filter *domain.SessionFilter) error {
	sql, args, err := r.builder.
		Update("session").
		Set("revoked", true).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": filter.ID}).
		ToSql()
	if err != nil {
		return apperrors.AutoSource(
			apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
				"failed to build update SQL query")).
			WithField("id", filter.ID.String()).
			WithDetails("Error occurred while building UPDATE query for revoke session")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(ctx, err, "session", map[string]any{
			"id": filter.ID.String(),
		})
	}

	return nil
}
