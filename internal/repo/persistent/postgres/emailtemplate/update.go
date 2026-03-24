package emailtemplate

import (
	"context"

	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/domain"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Update(ctx context.Context, t *domain.EmailTemplate) error {
	sql, args, err := r.builder.
		Update(table).
		Set("name", t.Name).
		Set("subject", t.Subject).
		Set("html_body", t.HtmlBody).
		Set("text_body", t.TextBody).
		Set("type", t.Type).
		Set("is_active", t.IsActive).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": t.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build update")
	}
	return r.pool.QueryRow(ctx, sql, args...).Scan(&t.UpdatedAt)
}
