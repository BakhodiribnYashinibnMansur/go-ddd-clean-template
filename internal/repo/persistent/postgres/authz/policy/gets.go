package policy

import (
	"context"
	"fmt"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Gets(ctx context.Context, filter *domain.PoliciesFilter) ([]*domain.Policy, int, error) {
	query := r.builder.Select("id", "permission_id", "effect", "priority", "active", "conditions", "created_at").From(tableName)
	countQuery := r.builder.Select("COUNT(*)").From(tableName)

	if filter.ID != nil {
		query = query.Where(squirrel.Eq{"id": *filter.ID})
		countQuery = countQuery.Where(squirrel.Eq{"id": *filter.ID})
	}
	if filter.PermissionID != nil {
		query = query.Where(squirrel.Eq{"permission_id": *filter.PermissionID})
		countQuery = countQuery.Where(squirrel.Eq{"permission_id": *filter.PermissionID})
	}
	if filter.Active != nil {
		query = query.Where(squirrel.Eq{"active": *filter.Active})
		countQuery = countQuery.Where(squirrel.Eq{"active": *filter.Active})
	}

	if filter.Pagination != nil {
		query = query.Limit(uint64(filter.Pagination.Limit)).Offset(uint64(filter.Pagination.Offset))
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build select query")
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		var p domain.Policy
		if err := rows.Scan(&p.ID, &p.PermissionID, &p.Effect, &p.Priority, &p.Active, &p.Conditions, &p.CreatedAt); err != nil {
			return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, fmt.Sprintf("failed to scan row: %v", err))
		}
		policies = append(policies, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}

	var count int
	countSql, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build count query")
	}
	err = r.pool.QueryRow(ctx, countSql, countArgs...).Scan(&count)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}

	return policies, count, nil
}
