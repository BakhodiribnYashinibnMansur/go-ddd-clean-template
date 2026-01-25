package session

import (
	"context"
	"time"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"

	"github.com/google/uuid"
)

func (r *Repo) Create(ctx context.Context, s *domain.Session) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	// Ensure timestamps are set
	now := time.Now()
	s.LastActivity = now

	query := r.builder.Insert(tableName).
		Columns(
			schema.SessionID,
			schema.SessionDeviceID,
			schema.SessionDeviceName,
			schema.SessionDeviceType,
			schema.SessionIPAddress,
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
		Values(
			s.ID,
			s.DeviceID,
			s.DeviceName,
			s.DeviceType,
			s.IPAddress,
			s.UserAgent,
			s.FCMToken,
			s.RefreshTokenHash,
			s.Data,
			s.UserID,
			s.Revoked,
			s.ExpiresAt,
			s.LastActivity,
			now,
			now,
		)

	sql, args, err := query.ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase,
			consts.ErrMsgFailedToBuildInsert)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
