package integration

import (
	"context"

	"gct/consts"
	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) UpdateAPIKey(ctx context.Context, apiKey *domain.APIKey) error {
	sql, args, err := r.builder.
		Update(tableAPIKeys).
		Set("name", apiKey.Name).
		Set("is_active", apiKey.IsActive).
		Set("expires_at", apiKey.ExpiresAt).
		Where(squirrel.Eq{"id": apiKey.ID}).
		Where(squirrel.Eq{"deleted_at": nil}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	err = r.pool.QueryRow(ctx, sql, args...).Scan(&apiKey.UpdatedAt)
	if err != nil {
		return apperrors.HandlePgError(err, tableAPIKeys, nil)
	}

	return nil
}
