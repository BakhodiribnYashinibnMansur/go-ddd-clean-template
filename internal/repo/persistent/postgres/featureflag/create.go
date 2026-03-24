package featureflag

import (
	"context"

	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/domain"
)

func (r *Repo) Create(ctx context.Context, f *domain.FeatureFlag) error {
	sql, args, err := r.builder.
		Insert(table).
		Columns("id", "key", "name", "type", "value", "description", "is_active").
		Values(f.ID, f.Key, f.Name, f.Type, f.Value, f.Description, f.IsActive).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build insert")
	}
	return r.pool.QueryRow(ctx, sql, args...).Scan(&f.CreatedAt, &f.UpdatedAt)
}
