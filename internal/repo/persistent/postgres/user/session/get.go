package session

import (
	"context"

	"github.com/Masterminds/squirrel"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) GetByID(ctx context.Context, filter *domain.SessionFilter) (*domain.Session, error) {
	sql, args, err := r.builder.
		Select("id, device_id, device_name, device_type, ip_address, user_agent, fcm_token, expires_at, last_activity, created_at, updated_at").
		From("session").
		Where(squirrel.Eq{"id": *filter.ID}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
			"failed to build select SQL query")
	}

	var s domain.Session
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&s.ID, &s.DeviceID, &s.DeviceName, &s.DeviceType, &s.IPAddress, &s.UserAgent, &s.FCMToken, &s.ExpiresAt, &s.LastActivity, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(ctx, err, "session", map[string]any{
			"id": filter.ID,
		})
	}

	return &s, nil
}
