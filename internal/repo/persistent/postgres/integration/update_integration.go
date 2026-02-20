package integration

import (
	"context"

	"gct/consts"
	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) UpdateIntegration(ctx context.Context, integration *domain.Integration) error {
	sql, args, err := r.builder.
		Update(tableIntegrations).
		Set("name", integration.Name).
		Set("description", integration.Description).
		Set("base_url", integration.BaseURL).
		Set("is_active", integration.IsActive).
		Set("config", integration.Config).
		Where(squirrel.Eq{"id": integration.ID}).
		Where(squirrel.Eq{"deleted_at": nil}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	err = r.pool.QueryRow(ctx, sql, args...).Scan(&integration.UpdatedAt)
	if err != nil {
		return apperrors.HandlePgError(err, tableIntegrations, nil)
	}

	return nil
}
