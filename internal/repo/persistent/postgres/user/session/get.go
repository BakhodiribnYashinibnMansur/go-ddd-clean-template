package session

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"github.com/Masterminds/squirrel"
)

func (r *Repo) Get(ctx context.Context, filter *domain.SessionFilter) (*domain.Session, error) {
	if filter == nil {
		return nil, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
			"session filter cannot be nil")
	}

	query := r.builder.Select("id, device_id, device_name, device_type, ip_address::text, user_agent, fcm_token, refresh_token_hash, data, user_id, revoked, expires_at, last_activity, created_at, updated_at").
		From("session")

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
		return nil, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
			"failed to build select SQL query")
	}

	var s domain.Session
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&s.ID, &s.DeviceID, &s.DeviceName, &s.DeviceType, &s.IPAddress, &s.UserAgent, &s.FCMToken, &s.RefreshTokenHash, &s.Data, &s.UserID, &s.Revoked, &s.ExpiresAt, &s.LastActivity, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(ctx, err, "session", map[string]any{
			"id": filter.ID,
		})
	}

	return &s, nil
}
