package session

import (
	"context"
	"time"

	"github.com/evrone/go-clean-template/internal/domain"
	apperrors "github.com/evrone/go-clean-template/pkg/errors"
	"github.com/google/uuid"
)

func (r *Repo) Create(ctx context.Context, s *domain.Session) error {
	query := r.builder.Insert("session").
		Columns(
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

	if s.ID != uuid.Nil {
		query = r.builder.Insert("session").
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
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return apperrors.AutoSource(
			apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
				"failed to build insert SQL query")).
			WithField("device_id", s.DeviceID.String()).
			WithDetails("Error occurred while building INSERT query for session")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(ctx, err, "session", map[string]any{
			"device_id": s.DeviceID.String(),
			"user_id":   s.UserID,
		})
	}

	return nil
}
