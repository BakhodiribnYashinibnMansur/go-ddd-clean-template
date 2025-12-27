package persistent

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/pkg/db/postgres"
	"github.com/google/uuid"
)

// SessionRepo -.
type SessionRepo struct {
	*postgres.Postgres
}

// NewSessionRepo -.
func NewSessionRepo(pg *postgres.Postgres) *SessionRepo {
	return &SessionRepo{pg}
}

// Create creates a new session.
func (r *SessionRepo) Create(ctx context.Context, s entity.Session) (entity.Session, error) {
	s.ID = uuid.New()
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	s.LastActivity = time.Now()

	sql, args, err := r.Builder.
		Insert("session").
		Columns("id", "turon_id", "device_id", "device_name", "device_type", "ip_address", "user_agent", "fcm_token", "expires_at", "last_activity", "created_at", "updated_at").
		Values(s.ID, s.TuronID, s.DeviceID, s.DeviceName, s.DeviceType, s.IPAddress, s.UserAgent, s.FCMToken, s.ExpiresAt, s.LastActivity, s.CreatedAt, s.UpdatedAt).
		ToSql()
	if err != nil {
		return entity.Session{}, fmt.Errorf("SessionRepo - Create - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return entity.Session{}, fmt.Errorf("SessionRepo - Create - r.Pool.Exec: %w", err)
	}

	return s, nil
}

// GetByID gets a session by ID.
func (r *SessionRepo) GetByID(ctx context.Context, id uuid.UUID) (entity.Session, error) {
	sql, args, err := r.Builder.
		Select("id", "turon_id", "device_id", "device_name", "device_type", "ip_address", "user_agent", "fcm_token", "expires_at", "last_activity", "created_at", "updated_at").
		From("session").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return entity.Session{}, fmt.Errorf("SessionRepo - GetByID - r.Builder: %w", err)
	}

	var s entity.Session
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&s.ID, &s.TuronID, &s.DeviceID, &s.DeviceName, &s.DeviceType, &s.IPAddress, &s.UserAgent, &s.FCMToken, &s.ExpiresAt, &s.LastActivity, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return entity.Session{}, fmt.Errorf("SessionRepo - GetByID - r.Pool.QueryRow: %w", err)
	}

	return s, nil
}

// GetByUserID gets all sessions for a user.
func (r *SessionRepo) GetByUserID(ctx context.Context, turonID int64) ([]entity.Session, error) {
	sql, args, err := r.Builder.
		Select("id", "turon_id", "device_id", "device_name", "device_type", "ip_address", "user_agent", "fcm_token", "expires_at", "last_activity", "created_at", "updated_at").
		From("session").
		Where("turon_id = ?", turonID).
		OrderBy("last_activity DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("SessionRepo - GetByUserID - r.Builder: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("SessionRepo - GetByUserID - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	var sessions []entity.Session
	for rows.Next() {
		var s entity.Session
		err = rows.Scan(
			&s.ID, &s.TuronID, &s.DeviceID, &s.DeviceName, &s.DeviceType, &s.IPAddress, &s.UserAgent, &s.FCMToken, &s.ExpiresAt, &s.LastActivity, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("SessionRepo - GetByUserID - rows.Scan: %w", err)
		}
		sessions = append(sessions, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("SessionRepo - GetByUserID - rows.Err: %w", err)
	}

	return sessions, nil
}

// GetByDeviceID gets a session by device ID and user ID.
func (r *SessionRepo) GetByDeviceID(ctx context.Context, turonID int64, deviceID uuid.UUID) (entity.Session, error) {
	sql, args, err := r.Builder.
		Select("id", "turon_id", "device_id", "device_name", "device_type", "ip_address", "user_agent", "fcm_token", "expires_at", "last_activity", "created_at", "updated_at").
		From("session").
		Where("turon_id = ? AND device_id = ?", turonID, deviceID).
		ToSql()
	if err != nil {
		return entity.Session{}, fmt.Errorf("SessionRepo - GetByDeviceID - r.Builder: %w", err)
	}

	var s entity.Session
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&s.ID, &s.TuronID, &s.DeviceID, &s.DeviceName, &s.DeviceType, &s.IPAddress, &s.UserAgent, &s.FCMToken, &s.ExpiresAt, &s.LastActivity, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return entity.Session{}, fmt.Errorf("SessionRepo - GetByDeviceID - r.Pool.QueryRow: %w", err)
	}

	return s, nil
}

// UpdateActivity updates session's last activity and FCM token.
func (r *SessionRepo) UpdateActivity(ctx context.Context, id uuid.UUID, fcmToken *string) error {
	sql, args, err := r.Builder.
		Update("session").
		Set("last_activity", time.Now()).
		Set("updated_at", time.Now()).
		Set("fcm_token", fcmToken).
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return fmt.Errorf("SessionRepo - UpdateActivity - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("SessionRepo - UpdateActivity - r.Pool.Exec: %w", err)
	}

	return nil
}

// Delete deletes a session by ID.
func (r *SessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := r.Builder.
		Delete("session").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return fmt.Errorf("SessionRepo - Delete - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("SessionRepo - Delete - r.Pool.Exec: %w", err)
	}

	return nil
}

// DeleteByUserID deletes all sessions for a user.
func (r *SessionRepo) DeleteByUserID(ctx context.Context, turonID int64) error {
	sql, args, err := r.Builder.
		Delete("session").
		Where("turon_id = ?", turonID).
		ToSql()
	if err != nil {
		return fmt.Errorf("SessionRepo - DeleteByUserID - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("SessionRepo - DeleteByUserID - r.Pool.Exec: %w", err)
	}

	return nil
}

// DeleteExpired deletes all expired sessions.
func (r *SessionRepo) DeleteExpired(ctx context.Context) error {
	sql, args, err := r.Builder.
		Delete("session").
		Where("expires_at < ?", time.Now()).
		ToSql()
	if err != nil {
		return fmt.Errorf("SessionRepo - DeleteExpired - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("SessionRepo - DeleteExpired - r.Pool.Exec: %w", err)
	}

	return nil
}
