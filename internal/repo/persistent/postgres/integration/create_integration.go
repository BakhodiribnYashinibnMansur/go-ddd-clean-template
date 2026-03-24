package integration

import (
	"context"

	"gct/internal/shared/domain/consts"
	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (r *Repo) CreateIntegration(ctx context.Context, integration *domain.Integration) error {
	sql, args, err := r.builder.
		Insert(tableIntegrations).
		Columns("id", "name", "description", "base_url", "is_active", "config").
		Values(integration.ID, integration.Name, integration.Description, integration.BaseURL, integration.IsActive, integration.Config).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	err = r.pool.QueryRow(ctx, sql, args...).Scan(&integration.CreatedAt, &integration.UpdatedAt)
	if err != nil {
		return apperrors.HandlePgError(err, tableIntegrations, nil)
	}

	return nil
}
