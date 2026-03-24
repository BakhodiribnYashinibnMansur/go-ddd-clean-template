package integration

import (
	"context"

	"gct/internal/shared/domain/consts"
	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) ListIntegrations(ctx context.Context, filter domain.IntegrationFilter) ([]domain.Integration, int64, error) {
	query := r.builder.
		Select("id", "name", "description", "base_url", "is_active", "config", "created_at", "updated_at", "deleted_at").
		From(tableIntegrations).
		Where(squirrel.Eq{"deleted_at": nil})

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where(squirrel.Or{
			squirrel.ILike{"name": searchPattern},
			squirrel.ILike{"description": searchPattern},
		})
	}

	if filter.IsActive != nil {
		query = query.Where(squirrel.Eq{"is_active": *filter.IsActive})
	}

	countSQL, countArgs, err := r.builder.
		Select("COUNT(*)").
		FromSelect(query, "filtered").
		ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var total int64
	err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableIntegrations, nil)
	}

	if filter.Limit > 0 {
		query = query.Limit(uint64(filter.Limit))
	}
	if filter.Offset > 0 {
		query = query.Offset(uint64(filter.Offset))
	}

	sql, args, err := query.OrderBy("created_at DESC").ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableIntegrations, nil)
	}
	defer rows.Close()

	var integrations []domain.Integration
	for rows.Next() {
		var integration domain.Integration
		if err := rows.Scan(
			&integration.ID,
			&integration.Name,
			&integration.Description,
			&integration.BaseURL,
			&integration.IsActive,
			&integration.Config,
			&integration.CreatedAt,
			&integration.UpdatedAt,
			&integration.DeletedAt,
		); err != nil {
			return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToScanRow)
		}
		integrations = append(integrations, integration)
	}

	return integrations, total, nil
}
