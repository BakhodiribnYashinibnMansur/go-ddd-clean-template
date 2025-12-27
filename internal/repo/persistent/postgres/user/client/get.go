package client

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/evrone/go-clean-template/internal/domain"
	"go.uber.org/zap"
)

func (r *Repo) Get(ctx context.Context, filter UserFilter) (domain.User, error) {
	r.logger.Info("UserRepo.Get started")

	qb := r.builder.
		Select("id, username, phone, password_hash, salt, created_at, updated_at, deleted_at, last_seen").
		From("users").
		Where("deleted_at = 0")

	if filter.ID != nil {
		qb = qb.Where(squirrel.Eq{"id": *filter.ID})
	}

	if filter.Phone != nil {
		qb = qb.Where(squirrel.Eq{"phone": *filter.Phone})
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		r.logger.Error("UserRepo.Get - r.builder", zap.Error(err))
		return domain.User{}, fmt.Errorf("UserRepo - Get - r.builder: %w", err)
	}

	var u domain.User
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&u.ID, &u.Username, &u.Phone, &u.PasswordHash, &u.Salt, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.LastSeen,
	)
	if err != nil {
		r.logger.Error("UserRepo.Get - r.psql.Pool.QueryRow", zap.Error(err))
		return domain.User{}, fmt.Errorf("UserRepo - Get - r.psql.Pool.QueryRow: %w", err)
	}

	username := ""
	if u.Username != nil {
		username = *u.Username
	}
	r.logger.Info("UserRepo.Get finished", zap.String("username", username))
	return u, nil
}
