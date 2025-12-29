package session

import (
	"context"
	"time"

	"github.com/google/uuid"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, s *domain.Session) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	query := r.builder.Insert("session").
		Columns(
			"id",
			"device_id",
			"device_name",
			"device_type",
			"ip_address",
			"user_agent",
			"fcm_token",
			"user_id",
			"company_id",
			"revoked",
			"expires_at",
			"last_activity",
			"created_at",
			"updated_at",
		).
		Values(
			s.ID,
			s.DeviceID,
			s.DeviceName,
			s.DeviceType,
			s.IPAddress,
			s.UserAgent,
			s.FCMToken,
			s.UserID,
			s.CompanyID,
			s.Revoked,
			s.ExpiresAt,
			s.LastActivity,
			time.Now(),
			time.Now(),
		)

	sql, args, err := query.ToSql()
	if err != nil {
		return apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
			"failed to build insert SQL query")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(ctx, err, "session", nil)
	}

	return nil
}
