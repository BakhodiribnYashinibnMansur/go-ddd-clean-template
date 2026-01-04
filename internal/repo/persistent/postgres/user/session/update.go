package session

import (
	"context"
	"time"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"github.com/Masterminds/squirrel"
)

func (r *Repo) Update(ctx context.Context, s *domain.Session) error {
	qb := r.builder.Update("session")

	if s.FCMToken != nil {
		qb = qb.Set("fcm_token", s.FCMToken)
	}
	if s.RefreshTokenHash != "" {
		qb = qb.Set("refresh_token_hash", s.RefreshTokenHash)
	}
	if s.Data != nil {
		qb = qb.Set("data", s.Data)
	}

	sql, args, err := qb.
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
	if filter == nil {
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
			"session filter cannot be nil")
	}

	query := r.builder.Update("session").
		Set("revoked", true).
		Set("updated_at", time.Now())

	if !filter.IsIDNull() {
		query = query.Where(squirrel.Eq{"id": *filter.ID})
	}
	if !filter.IsUserIDNull() {
		query = query.Where(squirrel.Eq{"user_id": *filter.UserID})
	}
	if !filter.IsRevokedNull() {
		query = query.Where(squirrel.Eq{"revoked": *filter.Revoked})
	}

	sql, args, err := query.ToSql()
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
