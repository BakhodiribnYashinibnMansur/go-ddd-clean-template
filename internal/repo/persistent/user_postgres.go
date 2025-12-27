package persistent

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/pkg/postgres"
)

type UserRepo struct {
	*postgres.Postgres
}

func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{pg}
}

func (r *UserRepo) Create(ctx context.Context, u entity.User) error {
	sql, args, err := r.Builder.
		Insert("users").
		Columns("username", "phone", "password_hash", "salt", "created_at", "updated_at", "deleted_at", "last_seen").
		Values(u.Username, u.Phone, u.PasswordHash, u.Salt, time.Now(), time.Now(), 0, u.LastSeen).
		ToSql()
	if err != nil {
		return fmt.Errorf("UserRepo - Create - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserRepo - Create - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (entity.User, error) {
	sql, args, err := r.Builder.
		Select("id, username, phone, password_hash, salt, created_at, updated_at, deleted_at, last_seen").
		From("users").
		Where("id = ? AND deleted_at = 0", id).
		ToSql()
	if err != nil {
		return entity.User{}, fmt.Errorf("UserRepo - GetByID - r.Builder: %w", err)
	}

	var u entity.User
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&u.ID, &u.Username, &u.Phone, &u.PasswordHash, &u.Salt, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.LastSeen,
	)
	if err != nil {
		return entity.User{}, fmt.Errorf("UserRepo - GetByID - r.Pool.QueryRow: %w", err)
	}

	return u, nil
}

func (r *UserRepo) GetByPhone(ctx context.Context, phone string) (entity.User, error) {
	sql, args, err := r.Builder.
		Select("id, username, phone, password_hash, salt, created_at, updated_at, deleted_at, last_seen").
		From("users").
		Where("phone = ? AND deleted_at = 0", phone).
		ToSql()
	if err != nil {
		return entity.User{}, fmt.Errorf("UserRepo - GetByPhone - r.Builder: %w", err)
	}

	var u entity.User
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(
		&u.ID, &u.Username, &u.Phone, &u.PasswordHash, &u.Salt, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.LastSeen,
	)
	if err != nil {
		return entity.User{}, fmt.Errorf("UserRepo - GetByPhone - r.Pool.QueryRow: %w", err)
	}

	return u, nil
}

func (r *UserRepo) Update(ctx context.Context, u entity.User) error {
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
		return fmt.Errorf("UserRepo - Update - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserRepo - Update - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id int64) error {
	sql, args, err := r.Builder.
		Update("users").
		Set("deleted_at", time.Now().Unix()).
		Set("updated_at", time.Now()).
		Where("id = ? AND deleted_at = 0", id).
		ToSql()
	if err != nil {
		return fmt.Errorf("UserRepo - Delete - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserRepo - Delete - r.Pool.Exec: %w", err)
	}

	return nil
}
