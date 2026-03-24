package relation

import (
	"context"
	"time"

	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) AddUser(ctx context.Context, relationID, userID uuid.UUID) error {
	sql, args, err := r.builder.
		Insert("user_relation").
		Columns("user_id", "relation_id", "created_at").
		Values(userID, relationID, time.Now()).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build insert query")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, "user_relation", nil)
	}

	return nil
}

func (r *Repo) RemoveUser(ctx context.Context, relationID, userID uuid.UUID) error {
	sql, args, err := r.builder.
		Delete("user_relation").
		Where(squirrel.Eq{"user_id": userID}).
		Where(squirrel.Eq{"relation_id": relationID}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build delete query")
	}

	tag, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, "user_relation", nil)
	}

	if tag.RowsAffected() == 0 {
		return apperrors.NewRepoError(apperrors.ErrRepoNotFound, "user relation not found")
	}

	return nil
}
