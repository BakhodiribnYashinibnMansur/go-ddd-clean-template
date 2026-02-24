package emailtemplate

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) ListLogs(ctx context.Context, filter domain.EmailLogFilter) ([]domain.EmailLog, int64, error) {
	q := r.builder.
		Select("id", "template_id", "to_email", "subject", "status", "error", "sent_at", "created_at").
		From(tableLog)

	if filter.TemplateID != "" {
		q = q.Where(squirrel.Eq{"template_id": filter.TemplateID})
	}
	if filter.Status != "" {
		q = q.Where(squirrel.Eq{"status": filter.Status})
	}

	countSQL, countArgs, _ := r.builder.Select("COUNT(*)").FromSelect(q, "sub").ToSql()
	var total int64
	if err := r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableLog, nil)
	}

	if filter.Limit > 0 {
		q = q.Limit(uint64(filter.Limit))
	}
	if filter.Offset > 0 {
		q = q.Offset(uint64(filter.Offset))
	}
	sql, args, err := q.OrderBy("created_at DESC").ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build list logs")
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableLog, nil)
	}
	defer rows.Close()

	var items []domain.EmailLog
	for rows.Next() {
		var l domain.EmailLog
		if err := rows.Scan(&l.ID, &l.TemplateID, &l.ToEmail, &l.Subject, &l.Status, &l.Error, &l.SentAt, &l.CreatedAt); err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableLog, nil)
		}
		items = append(items, l)
	}
	if items == nil {
		items = []domain.EmailLog{}
	}
	return items, total, nil
}
