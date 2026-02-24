package webhook

import (
	"context"
	"encoding/json"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Update(ctx context.Context, w *domain.Webhook) error {
	eventsJSON, err := json.Marshal(w.Events)
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "marshal events")
	}
	headersJSON, err := json.Marshal(w.Headers)
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "marshal headers")
	}
	sql, args, err := r.builder.
		Update(table).
		Set("name", w.Name).
		Set("url", w.URL).
		Set("secret", w.Secret).
		Set("events", eventsJSON).
		Set("headers", headersJSON).
		Set("is_active", w.IsActive).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": w.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build update")
	}
	return r.pool.QueryRow(ctx, sql, args...).Scan(&w.UpdatedAt)
}
