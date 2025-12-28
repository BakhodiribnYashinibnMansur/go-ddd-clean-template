package session

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/evrone/go-clean-template/internal/domain"
	apperrors "github.com/evrone/go-clean-template/pkg/errors"
)

func (r *Repo) Delete(ctx context.Context, filter *domain.SessionFilter) error {
	sql, args, err := r.builder.
		Delete("session").
		Where(squirrel.Eq{"id": filter.ID}).
		ToSql()
	if err != nil {
		return apperrors.AutoSource(
			apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase,
				"failed to build delete SQL query")).
			WithField("id", filter.ID.String()).
			WithDetails("Error occurred while building DELETE query for session")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(ctx, err, "session", map[string]any{
			"id": filter.ID.String(),
		})
	}

	return nil
}
