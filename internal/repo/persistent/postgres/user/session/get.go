package session

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (domain.Session, error) {
	r.logger.Info("SessionRepo.GetByID started", zap.String("id", id.String()))

	sql, args, err := r.builder.
		Select("id, device_id, device_name, device_type, ip_address, user_agent, fcm_token, expires_at, last_activity, created_at, updated_at").
		From("session").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		r.logger.Error("SessionRepo.GetByID - r.builder", zap.Error(err))
		return domain.Session{}, fmt.Errorf("SessionRepo - GetByID - r.builder: %w", err)
	}

	var s domain.Session
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&s.ID, &s.DeviceID, &s.DeviceName, &s.DeviceType, &s.IPAddress, &s.UserAgent, &s.FCMToken, &s.ExpiresAt, &s.LastActivity, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("SessionRepo.GetByID - r.pool.QueryRow", zap.Error(err))
		return domain.Session{}, fmt.Errorf("SessionRepo - GetByID - r.pool.QueryRow: %w", err)
	}

	r.logger.Info("SessionRepo.GetByID finished")
	return s, nil
}
