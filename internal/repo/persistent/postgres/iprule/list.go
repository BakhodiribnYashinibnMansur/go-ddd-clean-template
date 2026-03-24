package iprule

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) List(ctx context.Context, filter domain.IPRuleFilter) ([]domain.IPRule, int64, error) {
	q := r.builder.
		Select("id", "ip_address", "type", "reason", "is_active", "created_at", "updated_at").
		From(table)

	if filter.Search != "" {
		q = q.Where(squirrel.ILike{"ip_address": "%" + filter.Search + "%"})
	}
	if filter.Type != "" {
		q = q.Where(squirrel.Eq{"type": filter.Type})
	}
	if filter.IsActive != nil {
		q = q.Where(squirrel.Eq{"is_active": *filter.IsActive})
	}

	countSQL, countArgs, _ := r.builder.Select("COUNT(*)").FromSelect(q, "sub").ToSql()
	var total int64
	if err := r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, table, nil)
	}

	if filter.Limit > 0 {
		q = q.Limit(uint64(filter.Limit))
	}
	if filter.Offset > 0 {
		q = q.Offset(uint64(filter.Offset))
	}
	listSQL, args, err := q.OrderBy("created_at DESC").ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build list")
	}

	rows, err := r.pool.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, table, nil)
	}
	defer rows.Close()

	var items []domain.IPRule
	for rows.Next() {
		var rule domain.IPRule
		if err := rows.Scan(&rule.ID, &rule.IPAddress, &rule.Type, &rule.Reason, &rule.IsActive, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, 0, apperrors.HandlePgError(err, table, nil)
		}
		items = append(items, rule)
	}
	if items == nil {
		items = []domain.IPRule{}
	}
	return items, total, nil
}
