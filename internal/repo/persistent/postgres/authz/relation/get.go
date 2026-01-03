package relation

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Get(ctx context.Context, filter *domain.RelationFilter) (*domain.Relation, error) {
	query := r.builder.Select("id", "type", "name", "created_at").From("relation")

	if filter.ID != nil {
		query = query.Where(squirrel.Eq{"id": *filter.ID})
	}
	if filter.Type != nil {
		query = query.Where(squirrel.Eq{"type": *filter.Type})
	}
	if filter.Name != nil {
		query = query.Where(squirrel.Eq{"name": *filter.Name})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build select query")
	}

	var relation domain.Relation
	err = r.pool.QueryRow(ctx, sql, args...).Scan(&relation.ID, &relation.Type, &relation.Name, &relation.CreatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(ctx, err, "relation", nil)
	}

	return &relation, nil
}
