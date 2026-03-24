package emailtemplate

import (
	"context"

	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/domain"
)

func (r *Repo) Create(ctx context.Context, t *domain.EmailTemplate) error {
	sql, args, err := r.builder.
		Insert(table).
		Columns("id", "name", "subject", "html_body", "text_body", "type", "is_active").
		Values(t.ID, t.Name, t.Subject, t.HtmlBody, t.TextBody, t.Type, t.IsActive).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build insert")
	}
	return r.pool.QueryRow(ctx, sql, args...).Scan(&t.CreatedAt, &t.UpdatedAt)
}
