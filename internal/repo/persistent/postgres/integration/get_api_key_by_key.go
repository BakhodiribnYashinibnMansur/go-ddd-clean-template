package integration

import (
	"context"

	"gct/internal/shared/domain/consts"
	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) GetAPIKeyByKey(ctx context.Context, key string) (*domain.APIKey, error) {
	sql, args, err := r.builder.
		Select("id", "integration_id", "name", "key", "key_prefix", "is_active", "expires_at", "last_used_at", "created_at", "updated_at", "deleted_at").
		From(tableAPIKeys).
		Where(squirrel.Eq{"key": key}).
		Where(squirrel.Eq{"is_active": true}).
		Where(squirrel.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var apiKey domain.APIKey
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
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
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableAPIKeys, nil)
	}

	return &apiKey, nil
}
