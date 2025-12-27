package session

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (r *Repo) Update(ctx context.Context, s domain.Session) error {
	r.logger.Info("SessionRepo.Update started", zap.String("id", s.ID.String()))

	sql, args, err := r.builder.
		Update("session").
		Set("fcm_token", s.FCMToken).
		Set("revoked", s.Revoked).
		Set("last_activity", s.LastActivity).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": s.ID}).
		ToSql()
	if err != nil {
		r.logger.Error("SessionRepo.Update - r.builder", zap.Error(err))
		return fmt.Errorf("SessionRepo - Update - r.builder: %w", err)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		r.logger.Error("SessionRepo.Update - r.pool.Exec", zap.Error(err))
		return fmt.Errorf("SessionRepo - Update - r.pool.Exec: %w", err)
	}

	r.logger.Info("SessionRepo.Update finished")
	return nil
}

func (r *Repo) Revoke(ctx context.Context, id uuid.UUID) error {
	r.logger.Info("SessionRepo.Revoke started", zap.String("id", id.String()))

	sql, args, err := r.builder.
		Update("session").
		Set("revoked", true).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("SessionRepo - Revoke - r.builder: %w", err)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("SessionRepo - Revoke - r.pool.Exec: %w", err)
	}

	r.logger.Info("SessionRepo.Revoke finished")
	return nil
}
