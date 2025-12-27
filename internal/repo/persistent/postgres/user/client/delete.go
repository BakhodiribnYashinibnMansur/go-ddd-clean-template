package client

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

func (r *Repo) Delete(ctx context.Context, id int64) error {
	r.logger.Info("UserRepo.Delete started", zap.Int64("id", id))

	sql, args, err := r.builder.
		Update("users").
		Set("deleted_at", time.Now().Unix()).
		Set("updated_at", time.Now()).
		Where("id = ? AND deleted_at = 0", id).
		ToSql()
	if err != nil {
		r.logger.Error("UserRepo.Delete - r.builder", zap.Error(err))
		return fmt.Errorf("UserRepo - Delete - r.builder: %w", err)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		r.logger.Error("UserRepo.Delete - r.pool.Exec", zap.Error(err))
		return fmt.Errorf("UserRepo - Delete - r.pool.Exec: %w", err)
	}

	r.logger.Info("UserRepo.Delete finished")
	return nil
}
