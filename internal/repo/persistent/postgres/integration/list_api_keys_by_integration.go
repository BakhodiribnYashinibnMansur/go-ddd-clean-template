package integration

import (
	"context"

	"gct/consts"
	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) ListAPIKeysByIntegration(ctx context.Context, integrationID uuid.UUID) ([]domain.APIKey, error) {
	sql, args, err := r.builder.
		Select("id", "integration_id", "name", "key", "key_prefix", "is_active", "expires_at", "last_used_at", "created_at", "updated_at", "deleted_at").
		From(tableAPIKeys).
		Where(squirrel.Eq{"integration_id": integrationID}).
		Where(squirrel.Eq{"deleted_at": nil}).
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableAPIKeys, nil)
	}
	defer rows.Close()

	var apiKeys []domain.APIKey
	for rows.Next() {
		var apiKey domain.APIKey
		if err := rows.Scan(
			&apiKey.ID,
			&apiKey.IntegrationID,
			&apiKey.Name,
			&apiKey.Key,
			&apiKey.KeyPrefix,
			&apiKey.IsActive,
			&apiKey.ExpiresAt,
			&apiKey.LastUsedAt,
			&apiKey.CreatedAt,
			&apiKey.UpdatedAt,
			&apiKey.DeletedAt,
		); err != nil {
			return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToScanRow)
		}
		apiKeys = append(apiKeys, apiKey)
	}

	return apiKeys, nil
}
