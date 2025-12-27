package client

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/evrone/go-clean-template/internal/domain"
	"go.uber.org/zap"
)

func (r *Repo) Gets(ctx context.Context, filter UserListFilter) ([]domain.User, int, error) {
	r.logger.Info("UserRepo.Gets started")

	// Base query
	qb := r.builder.
		Select("id, username, phone, password_hash, salt, created_at, updated_at, deleted_at, last_seen").
		From("users").
		Where("deleted_at = 0")

	// Apply filters
	if filter.ID != nil {
		qb = qb.Where(squirrel.Eq{"id": *filter.ID})
	}

	if filter.Phone != nil {
		qb = qb.Where(squirrel.Eq{"phone": *filter.Phone})
	}

	// Count query
	countQb := r.builder.Select("COUNT(*)").From("users").Where("deleted_at = 0")
	if filter.ID != nil {
		countQb = countQb.Where(squirrel.Eq{"id": *filter.ID})
	}
	if filter.Phone != nil {
		countQb = countQb.Where(squirrel.Eq{"phone": *filter.Phone})
	}

	countSql, countArgs, err := countQb.ToSql()
	if err != nil {
		r.logger.Error("UserRepo.Gets - count r.builder", zap.Error(err))
		return nil, 0, fmt.Errorf("UserRepo - Gets - count r.builder: %w", err)
	}

	var count int
	err = r.pool.QueryRow(ctx, countSql, countArgs...).Scan(&count)
	if err != nil {
		r.logger.Error("UserRepo.Gets - r.pool.QueryRow count", zap.Error(err))
		return nil, 0, fmt.Errorf("UserRepo - Gets - r.pool.QueryRow count: %w", err)
	}

	// Apply pagination
	if filter.Limit > 0 {
		qb = qb.Limit(uint64(filter.Limit))
	}
	if filter.Offset > 0 {
		qb = qb.Offset(uint64(filter.Offset))
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		r.logger.Error("UserRepo.Gets - r.builder", zap.Error(err))
		return nil, 0, fmt.Errorf("UserRepo - Gets - r.builder: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		r.logger.Error("UserRepo.Gets - r.pool.Query", zap.Error(err))
		return nil, 0, fmt.Errorf("UserRepo - Gets - r.pool.Query: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		err = rows.Scan(
			&u.ID, &u.Username, &u.Phone, &u.PasswordHash, &u.Salt, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.LastSeen,
		)
		if err != nil {
			r.logger.Error("UserRepo.Gets - rows.Scan", zap.Error(err))
			return nil, 0, fmt.Errorf("UserRepo - Gets - rows.Scan: %w", err)
		}
		users = append(users, u)
	}

	r.logger.Info("UserRepo.Gets finished", zap.Int("count", count), zap.Int("returned", len(users)))
	return users, count, nil
}
