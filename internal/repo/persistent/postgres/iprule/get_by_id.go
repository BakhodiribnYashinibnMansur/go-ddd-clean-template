package iprule

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*domain.IPRule, error) {
	sql, args, err := r.builder.
		Select("id", "ip_address", "type", "reason", "is_active", "created_at", "updated_at").
		From(table).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build select")
	}
	var rule domain.IPRule
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&rule.ID, &rule.IPAddress, &rule.Type, &rule.Reason, &rule.IsActive, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, table, nil)
	}
	return &rule, nil
}
