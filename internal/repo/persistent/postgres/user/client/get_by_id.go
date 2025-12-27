package client

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/internal/domain"
	"go.uber.org/zap"
)

func (r *UserRepo) GetByID(ctx context.Context, id int64) (domain.User, error) {
	r.logger.Info("UserRepo.GetByID started", zap.Int64("id", id))

	sql, args, err := r.Builder.
		Select("id, username, phone, password_hash, salt, created_at, updated_at, deleted_at, last_seen").
		From("users").
		Where("id = ? AND deleted_at = 0", id).
		ToSql()
	if err != nil {
		r.logger.Error("UserRepo.GetByID - r.Builder", zap.Error(err))
		return domain.User{}, fmt.Errorf("UserRepo - GetByID - r.Builder: %w", err)
	}

	var u domain.User
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&u.ID, &u.Username, &u.Phone, &u.PasswordHash, &u.Salt, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.LastSeen,
	)
	if err != nil {
		r.logger.Error("UserRepo.GetByID - r.Pool.QueryRow", zap.Error(err))
		return domain.User{}, fmt.Errorf("UserRepo - GetByID - r.Pool.QueryRow: %w", err)
	}

	username := ""
	if u.Username != nil {
		username = *u.Username
	}
	r.logger.Info("UserRepo.GetByID finished", zap.String("username", username))
	return u, nil
}
