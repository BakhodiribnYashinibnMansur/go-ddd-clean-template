package session

import (
	"context"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Get(ctx context.Context, filter *domain.SessionFilter) (*domain.Session, error) {
	if filter == nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase,
			consts.ErrMsgInvalidInput)
	}

	query := r.builder.Select(
		schema.SessionID,
		schema.SessionDeviceID,
		schema.SessionDeviceName,
		schema.SessionDeviceType,
		schema.SessionIPAddress+"::text",
		schema.SessionUserAgent,
		schema.SessionFCMToken,
		schema.SessionRefreshTokenHash,
		schema.SessionData,
		schema.SessionUserID,
		schema.SessionRevoked,
		schema.SessionExpiresAt,
		schema.SessionLastActivity,
		schema.SessionCreatedAt,
		schema.SessionUpdatedAt,
	).
		From(tableName)

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
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase,
			consts.ErrMsgFailedToBuildQuery)
	}

	var s domain.Session
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&s.ID, &s.DeviceID, &s.DeviceName, &s.DeviceType, &s.IPAddress, &s.UserAgent, &s.FCMToken, &s.RefreshTokenHash, &s.Data, &s.UserID, &s.Revoked, &s.ExpiresAt, &s.LastActivity, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, map[string]any{
			"id": filter.ID,
		})
	}

	return &s, nil
}
