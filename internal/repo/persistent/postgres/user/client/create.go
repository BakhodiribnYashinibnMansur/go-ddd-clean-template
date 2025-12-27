package client

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/domain"
	"go.uber.org/zap"
)

func (r *Repo) Create(ctx context.Context, u domain.User) error {
	username := ""
	if u.Username != nil {
		username = *u.Username
	}
	r.logger.Info("UserRepo.Create started", zap.String("username", username))

	sql, args, err := r.builder.
		Insert("users").
		Columns("username", "phone", "password_hash", "salt", "created_at", "updated_at", "deleted_at", "last_seen").
		Values(u.Username, u.Phone, u.PasswordHash, u.Salt, time.Now(), time.Now(), 0, u.LastSeen).
		ToSql()
	if err != nil {
		r.logger.Error("UserRepo.Create - r.psql.Builder", zap.Error(err))
		return fmt.Errorf("UserRepo - Create - r.psql.Builder: %w", err)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		r.logger.Error("UserRepo.Create - r.pool.Exec", zap.Error(err))
		return fmt.Errorf("UserRepo - Create - r.pool.Exec: %w", err)
	}

	r.logger.Info("UserRepo.Create finished")
	return nil
}
