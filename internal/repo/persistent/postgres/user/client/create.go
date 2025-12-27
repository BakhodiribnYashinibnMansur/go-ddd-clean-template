package client

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/domain"
	"go.uber.org/zap"
)

func (r *UserRepo) Create(ctx context.Context, u domain.User) error {
	username := ""
	if u.Username != nil {
		username = *u.Username
	}
	r.logger.Info("UserRepo.Create started", zap.String("username", username))

	sql, args, err := r.Builder.
		Insert("users").
		Columns("username", "phone", "password_hash", "salt", "created_at", "updated_at", "deleted_at", "last_seen").
		Values(u.Username, u.Phone, u.PasswordHash, u.Salt, time.Now(), time.Now(), 0, u.LastSeen).
		ToSql()
	if err != nil {
		r.logger.Error("UserRepo.Create - r.Builder", zap.Error(err))
		return fmt.Errorf("UserRepo - Create - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		r.logger.Error("UserRepo.Create - r.Pool.Exec", zap.Error(err))
		return fmt.Errorf("UserRepo - Create - r.Pool.Exec: %w", err)
	}

	r.logger.Info("UserRepo.Create finished")
	return nil
}
