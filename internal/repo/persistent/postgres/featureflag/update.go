package featureflag

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Update(ctx context.Context, f *domain.FeatureFlag) error {
	sql, args, err := r.builder.
		Update(table).
		Set("name", f.Name).
		Set("type", f.Type).
		Set("value", f.Value).
		Set("description", f.Description).
		Set("is_active", f.IsActive).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": f.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build update")
	}
	return r.pool.QueryRow(ctx, sql, args...).Scan(&f.UpdatedAt)
}
