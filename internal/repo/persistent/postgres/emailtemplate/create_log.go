package emailtemplate

import (
	"context"

	apperrors "gct/pkg/errors"
	"gct/internal/domain"
)

func (r *Repo) CreateLog(ctx context.Context, l *domain.EmailLog) error {
	sql, args, err := r.builder.
		Insert(tableLog).
		Columns("id", "template_id", "to_email", "subject", "status", "error", "sent_at").
		Values(l.ID, l.TemplateID, l.ToEmail, l.Subject, l.Status, l.Error, l.SentAt).
		Suffix("RETURNING created_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build insert log")
	}
	return r.pool.QueryRow(ctx, sql, args...).Scan(&l.CreatedAt)
}
