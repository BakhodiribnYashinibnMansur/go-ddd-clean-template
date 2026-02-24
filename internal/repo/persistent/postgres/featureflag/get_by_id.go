package featureflag

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*domain.FeatureFlag, error) {
	sql, args, err := r.builder.
		Select("id", "key", "name", "type", "value", "description", "is_active", "created_at", "updated_at", "deleted_at").
		From(table).
		Where(squirrel.Eq{"id": id, "deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build select")
	}
	var f domain.FeatureFlag
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&f.ID, &f.Key, &f.Name, &f.Type, &f.Value, &f.Description, &f.IsActive,
		&f.CreatedAt, &f.UpdatedAt, &f.DeletedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, table, nil)
	}
	return &f, nil
}
