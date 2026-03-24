package iprule

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Update(ctx context.Context, rule *domain.IPRule) error {
	sql, args, err := r.builder.
		Update(table).
		Set("ip_address", rule.IPAddress).
		Set("type", rule.Type).
		Set("reason", rule.Reason).
		Set("is_active", rule.IsActive).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": rule.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build update")
	}
	return r.pool.QueryRow(ctx, sql, args...).Scan(&rule.UpdatedAt)
}
