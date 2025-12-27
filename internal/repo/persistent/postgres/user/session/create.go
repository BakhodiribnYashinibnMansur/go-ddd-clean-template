package session

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/domain"
	"go.uber.org/zap"
)

func (r *Repo) Create(ctx context.Context, s domain.Session) error {
	r.logger.Info("SessionRepo.Create started", zap.String("device_id", s.DeviceID.String()))

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

	if s.ID.String() != "00000000-0000-0000-0000-000000000000" {
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
		r.logger.Error("SessionRepo.Create - r.builder", zap.Error(err))
		return fmt.Errorf("SessionRepo - Create - r.builder: %w", err)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		r.logger.Error("SessionRepo.Create - r.pool.Exec", zap.Error(err))
		return fmt.Errorf("SessionRepo - Create - r.pool.Exec: %w", err)
	}

	r.logger.Info("SessionRepo.Create finished")
	return nil
}
