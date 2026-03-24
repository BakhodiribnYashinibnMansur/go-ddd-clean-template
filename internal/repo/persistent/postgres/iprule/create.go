package iprule

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (r *Repo) Create(ctx context.Context, rule *domain.IPRule) error {
	sql, args, err := r.builder.
		Insert(table).
		Columns("id", "ip_address", "type", "reason", "is_active").
		Values(rule.ID, rule.IPAddress, rule.Type, rule.Reason, rule.IsActive).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build insert")
	}
	return r.pool.QueryRow(ctx, sql, args...).Scan(&rule.CreatedAt, &rule.UpdatedAt)
}
