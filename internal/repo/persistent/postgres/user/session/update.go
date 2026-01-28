package session

import (
	"context"
	"time"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Update(ctx context.Context, s *domain.Session) error {
	qb := r.builder.Update(tableName)

	if s.FCMToken != nil {
		qb = qb.Set(schema.SessionFCMToken, s.FCMToken)
	}
	if s.RefreshTokenHash != "" {
		qb = qb.Set(schema.SessionRefreshTokenHash, s.RefreshTokenHash)
	}
	if s.Data != nil {
		qb = qb.Set(schema.SessionData, s.Data)
	}

	sql, args, err := qb.
		Set(schema.SessionRevoked, s.Revoked).
		Set(schema.SessionLastActivity, s.LastActivity).
		Set(schema.SessionUpdatedAt, time.Now()).
		Where(squirrel.Eq{schema.SessionID: s.ID}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase,
			consts.ErrMsgFailedToBuildUpdate)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

func (r *Repo) Revoke(ctx context.Context, filter *domain.SessionFilter) error {
	if filter == nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase,
			consts.ErrMsgInvalidInput)
	}

	query := r.builder.Update(tableName).
		Set(schema.SessionRevoked, true).
		Set(schema.SessionUpdatedAt, time.Now())

	if !filter.IsIDNull() {
		query = query.Where(squirrel.Eq{schema.SessionID: *filter.ID})
	}
	if !filter.IsUserIDNull() {
		query = query.Where(squirrel.Eq{schema.SessionUserID: *filter.UserID})
	}
	if !filter.IsRevokedNull() {
		query = query.Where(squirrel.Eq{schema.SessionRevoked: *filter.Revoked})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase,
			consts.ErrMsgFailedToBuildUpdate)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
