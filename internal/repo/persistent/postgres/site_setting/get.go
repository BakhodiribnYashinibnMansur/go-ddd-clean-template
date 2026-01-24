package sitesetting

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Get(ctx context.Context, filter *domain.SiteSettingFilter) (*domain.SiteSetting, error) {
	query := r.builder.Select(
		"id",
		"key",
		"value",
		"value_type",
		"category",
		"description",
		"is_public",
		"created_at",
		"updated_at",
	).From(tableName)

	if filter.ID != nil {
		query = query.Where(squirrel.Eq{"id": filter.ID})
	}
	if filter.Key != nil {
		query = query.Where(squirrel.Eq{"key": filter.Key})
	}
	if filter.Category != nil {
		query = query.Where(squirrel.Eq{"category": filter.Category})
	}
	if filter.IsPublic != nil {
		query = query.Where(squirrel.Eq{"is_public": filter.IsPublic})
	}

	query = query.Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build query")
	}

	var setting domain.SiteSetting
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&setting.ID,
		&setting.Key,
		&setting.Value,
		&setting.ValueType,
		&setting.Category,
		&setting.Description,
		&setting.IsPublic,
		&setting.CreatedAt,
		&setting.UpdatedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(ctx, err, tableName, nil)
	}

	return &setting, nil
}
