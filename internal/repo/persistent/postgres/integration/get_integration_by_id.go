package integration

import (
	"context"

	"gct/consts"
	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) GetIntegrationByID(ctx context.Context, id uuid.UUID) (*domain.Integration, error) {
	sql, args, err := r.builder.
		Select("id", "name", "description", "base_url", "is_active", "config", "created_at", "updated_at", "deleted_at").
		From(tableIntegrations).
		Where(squirrel.Eq{"id": id}).
		Where(squirrel.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var integration domain.Integration
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&integration.ID,
		&integration.Name,
		&integration.Description,
		&integration.BaseURL,
		&integration.IsActive,
		&integration.Config,
		&integration.CreatedAt,
		&integration.UpdatedAt,
		&integration.DeletedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableIntegrations, nil)
	}

	return &integration, nil
}
