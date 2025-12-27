package session

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (r *Repo) Delete(ctx context.Context, id uuid.UUID) error {
	r.logger.Info("SessionRepo.Delete started", zap.String("id", id.String()))

	sql, args, err := r.builder.
		Delete("session").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		r.logger.Error("SessionRepo.Delete - r.builder", zap.Error(err))
		return fmt.Errorf("SessionRepo - Delete - r.builder: %w", err)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		r.logger.Error("SessionRepo.Delete - r.pool.Exec", zap.Error(err))
		return fmt.Errorf("SessionRepo - Delete - r.pool.Exec: %w", err)
	}

	r.logger.Info("SessionRepo.Delete finished")
	return nil
}
