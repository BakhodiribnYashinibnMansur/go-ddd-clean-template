package relation

import (
	"context"
	"fmt"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"github.com/Masterminds/squirrel"
)

func (r *Repo) Gets(ctx context.Context, filter *domain.RelationsFilter) ([]*domain.Relation, int, error) {
	query := r.builder.Select("id", "type", "name", "created_at").From("relation")
	countQuery := r.builder.Select("COUNT(*)").From("relation")

	if filter.ID != nil {
		query = query.Where(squirrel.Eq{"id": *filter.ID})
		countQuery = countQuery.Where(squirrel.Eq{"id": *filter.ID})
	}
	if filter.Type != nil {
		query = query.Where(squirrel.Eq{"type": *filter.Type})
		countQuery = countQuery.Where(squirrel.Eq{"type": *filter.Type})
	}
	if filter.Name != nil {
		query = query.Where(squirrel.Eq{"name": *filter.Name})
		countQuery = countQuery.Where(squirrel.Eq{"name": *filter.Name})
	}

	if filter.Pagination != nil {
		query = query.Limit(uint64(filter.Pagination.Limit)).Offset(uint64(filter.Pagination.Offset))
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build select query")
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(ctx, err, "relation", nil)
	}
	defer rows.Close()

	var relations []*domain.Relation
	for rows.Next() {
		var relation domain.Relation
		if err := rows.Scan(&relation.ID, &relation.Type, &relation.Name, &relation.CreatedAt); err != nil {
			return nil, 0, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, fmt.Sprintf("failed to scan row: %v", err))
		}
		relations = append(relations, &relation)
	}

	// Count
	var count int
	countSql, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(ctx, apperrors.ErrRepoDatabase, "failed to build count query")
	}
	err = r.pool.QueryRow(ctx, countSql, countArgs...).Scan(&count)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(ctx, err, "relation", nil)
	}

	return relations, count, nil
}
