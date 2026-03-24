package sitesetting

import (
	"context"

	"gct/internal/shared/domain/consts"
	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Gets(ctx context.Context, filter *domain.SiteSettingsFilter) ([]*domain.SiteSetting, int, error) {
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

	// Build WHERE conditions
	if filter.Key != nil {
		query = query.Where(squirrel.Eq{"key": filter.Key})
	}
	if filter.Category != nil {
		query = query.Where(squirrel.Eq{"category": filter.Category})
	}
	if filter.IsPublic != nil {
		query = query.Where(squirrel.Eq{"is_public": filter.IsPublic})
	}

	// Count query
	countQuery := r.builder.Select("COUNT(*)").From(tableName)
	if filter.Key != nil {
		countQuery = countQuery.Where(squirrel.Eq{"key": filter.Key})
	}
	if filter.Category != nil {
		countQuery = countQuery.Where(squirrel.Eq{"category": filter.Category})
	}
	if filter.IsPublic != nil {
		countQuery = countQuery.Where(squirrel.Eq{"is_public": filter.IsPublic})
	}

	sql, args, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var count int
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}

	// Apply pagination
	if filter.Pagination != nil {
		if filter.Pagination.Limit > 0 {
			query = query.Limit(uint64(filter.Pagination.Limit))
		}
		if filter.Pagination.Offset > 0 {
			query = query.Offset(uint64(filter.Pagination.Offset))
		}
	}

	// Order by category, then key
	query = query.OrderBy("category"+" "+consts.SQLOrderAsc, "key"+" "+consts.SQLOrderAsc)

	sql, args, err = query.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}
	defer rows.Close()

	var settings []*domain.SiteSetting
	for rows.Next() {
		var s domain.SiteSetting
		err = rows.Scan(
			&s.ID,
			&s.Key,
			&s.Value,
			&s.ValueType,
			&s.Category,
			&s.Description,
			&s.IsPublic,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToScanRow)
		}
		settings = append(settings, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}

	return settings, count, nil
}
