package emailtemplate

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) List(ctx context.Context, filter domain.EmailTemplateFilter) ([]domain.EmailTemplate, int64, error) {
	q := r.builder.
		Select("id", "name", "subject", "html_body", "text_body", "type", "is_active", "created_at", "updated_at").
		From(table)

	if filter.Search != "" {
		q = q.Where(squirrel.ILike{"name": "%" + filter.Search + "%"})
	}
	if filter.Type != "" {
		q = q.Where(squirrel.Eq{"type": filter.Type})
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
	sql, args, err := q.OrderBy("created_at DESC").ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build list")
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, table, nil)
	}
	defer rows.Close()

	var items []domain.EmailTemplate
	for rows.Next() {
		var t domain.EmailTemplate
		if err := rows.Scan(&t.ID, &t.Name, &t.Subject, &t.HtmlBody, &t.TextBody, &t.Type, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, apperrors.HandlePgError(err, table, nil)
		}
		items = append(items, t)
	}
	if items == nil {
		items = []domain.EmailTemplate{}
	}
	return items, total, nil
}
