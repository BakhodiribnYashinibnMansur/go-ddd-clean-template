package announcement

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) List(ctx context.Context, filter domain.AnnouncementFilter) ([]domain.Announcement, int64, error) {
	q := r.builder.
		Select("id", "title", "content", "type", "is_active", "starts_at", "ends_at", "created_at", "updated_at").
		From(table)

	if filter.Search != "" {
		q = q.Where(squirrel.ILike{"title": "%" + filter.Search + "%"})
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

	var items []domain.Announcement
	for rows.Next() {
		var a domain.Announcement
		if err := rows.Scan(&a.ID, &a.Title, &a.Content, &a.Type, &a.IsActive, &a.StartsAt, &a.EndsAt, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, apperrors.HandlePgError(err, table, nil)
		}
		items = append(items, a)
	}
	if items == nil {
		items = []domain.Announcement{}
	}
	return items, total, nil
}
