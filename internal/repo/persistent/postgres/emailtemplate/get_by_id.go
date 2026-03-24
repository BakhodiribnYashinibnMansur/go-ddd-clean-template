package emailtemplate

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) GetByID(ctx context.Context, id string) (*domain.EmailTemplate, error) {
	sql, args, err := r.builder.
		Select("id", "name", "subject", "html_body", "text_body", "type", "is_active", "created_at", "updated_at").
		From(table).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build select")
	}
	var t domain.EmailTemplate
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&t.ID, &t.Name, &t.Subject, &t.HtmlBody, &t.TextBody, &t.Type, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, table, nil)
	}
	return &t, nil
}
