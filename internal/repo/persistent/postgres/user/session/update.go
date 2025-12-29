package session

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
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
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
			"failed to build update SQL query")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(ctx, err, "session", nil)
	}

	return nil
}

func (r *Repo) Revoke(ctx context.Context, filter *domain.SessionFilter) error {
	sql, args, err := r.builder.
		Update("session").
		Set("revoked", true).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": *filter.ID}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
			"failed to build update SQL query")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(ctx, err, "session", nil)
	}

	return nil
}
