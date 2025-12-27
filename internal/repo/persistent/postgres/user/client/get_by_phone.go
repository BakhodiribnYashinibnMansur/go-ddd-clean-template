package client

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/internal/domain"
	"go.uber.org/zap"
)

func (r *UserRepo) GetByPhone(ctx context.Context, phone string) (domain.User, error) {
	r.logger.Info("UserRepo.GetByPhone started", zap.String("phone", phone))

	sql, args, err := r.Builder.
		Select("id, username, phone, password_hash, salt, created_at, updated_at, deleted_at, last_seen").
		From("users").
		Where("phone = ? AND deleted_at = 0", phone).
		ToSql()
	if err != nil {
		r.logger.Error("UserRepo.GetByPhone - r.Builder", zap.Error(err))
		return domain.User{}, fmt.Errorf("UserRepo - GetByPhone - r.Builder: %w", err)
	}

	var u domain.User
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&u.ID, &u.Username, &u.Phone, &u.PasswordHash, &u.Salt, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.LastSeen,
	)
	if err != nil {
		r.logger.Error("UserRepo.GetByPhone - r.Pool.QueryRow", zap.Error(err))
		return domain.User{}, fmt.Errorf("UserRepo - GetByPhone - r.Pool.QueryRow: %w", err)
	}

	username := ""
	if u.Username != nil {
		username = *u.Username
	}
	r.logger.Info("UserRepo.GetByPhone finished", zap.String("username", username))
	return u, nil
}
