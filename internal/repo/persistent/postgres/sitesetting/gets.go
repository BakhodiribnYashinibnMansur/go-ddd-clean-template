package sitesetting

import (
	"context"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Gets(ctx context.Context, filter *domain.SiteSettingsFilter) ([]*domain.SiteSetting, int, error) {
	query := r.builder.Select(
		schema.SiteSettingID,
		schema.SiteSettingKey,
		schema.SiteSettingValue,
		schema.SiteSettingValueType,
		schema.SiteSettingCategory,
		schema.SiteSettingDescription,
		schema.SiteSettingIsPublic,
		schema.SiteSettingCreatedAt,
		schema.SiteSettingUpdatedAt,
	).From(tableName)

	// Build WHERE conditions
	if filter.Key != nil {
		query = query.Where(squirrel.Eq{schema.SiteSettingKey: filter.Key})
	}
	if filter.Category != nil {
		query = query.Where(squirrel.Eq{schema.SiteSettingCategory: filter.Category})
	}
	if filter.IsPublic != nil {
		query = query.Where(squirrel.Eq{schema.SiteSettingIsPublic: filter.IsPublic})
	}

	// Count query
	countQuery := r.builder.Select("COUNT(*)").From(tableName)
	if filter.Key != nil {
		countQuery = countQuery.Where(squirrel.Eq{schema.SiteSettingKey: filter.Key})
	}
	if filter.Category != nil {
		countQuery = countQuery.Where(squirrel.Eq{schema.SiteSettingCategory: filter.Category})
	}
	if filter.IsPublic != nil {
		countQuery = countQuery.Where(squirrel.Eq{schema.SiteSettingIsPublic: filter.IsPublic})
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
	query = query.OrderBy(schema.SiteSettingCategory+" "+consts.SQLOrderAsc, schema.SiteSettingKey+" "+consts.SQLOrderAsc)

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
