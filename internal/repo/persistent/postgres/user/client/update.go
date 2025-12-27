package client

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/domain"
	"go.uber.org/zap"
)

func (r *UserRepo) Update(ctx context.Context, u domain.User) error {
	username := ""
	if u.Username != nil {
		username = *u.Username
	}
	r.logger.Info("UserRepo.Update started", zap.Int64("id", u.ID), zap.String("username", username))

	sql, args, err := r.Builder.
		Update("users").
		Set("username", u.Username).
		Set("phone", u.Phone).
		Set("password_hash", u.PasswordHash).
		Set("salt", u.Salt).
		Set("updated_at", time.Now()).
		Set("last_seen", u.LastSeen).
		Where("id = ? AND deleted_at = 0", u.ID).
		ToSql()
	if err != nil {
		r.logger.Error("UserRepo.Update - r.Builder", zap.Error(err))
		return fmt.Errorf("UserRepo - Update - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		r.logger.Error("UserRepo.Update - r.Pool.Exec", zap.Error(err))
		return fmt.Errorf("UserRepo - Update - r.Pool.Exec: %w", err)
	}

	r.logger.Info("UserRepo.Update finished")
	return nil
}
