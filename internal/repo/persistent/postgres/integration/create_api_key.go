package integration

import (
	"context"

	"gct/internal/shared/domain/consts"
	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"
)

func (r *Repo) CreateAPIKey(ctx context.Context, apiKey *domain.APIKey) error {
	sql, args, err := r.builder.
		Insert(tableAPIKeys).
		Columns("id", "integration_id", "name", "key", "key_prefix", "is_active", "expires_at").
		Values(apiKey.ID, apiKey.IntegrationID, apiKey.Name, apiKey.Key, apiKey.KeyPrefix, apiKey.IsActive, apiKey.ExpiresAt).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	err = r.pool.QueryRow(ctx, sql, args...).Scan(&apiKey.CreatedAt, &apiKey.UpdatedAt)
	if err != nil {
		return apperrors.HandlePgError(err, tableAPIKeys, nil)
	}

	return nil
}
