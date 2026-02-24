package webhook

import (
	"context"
	"encoding/json"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, w *domain.Webhook) error {
	eventsJSON, err := json.Marshal(w.Events)
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "marshal events")
	}
	headersJSON, err := json.Marshal(w.Headers)
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "marshal headers")
	}
	sql, args, err := r.builder.
		Insert(table).
		Columns("id", "name", "url", "secret", "events", "headers", "is_active").
		Values(w.ID, w.Name, w.URL, w.Secret, eventsJSON, headersJSON, w.IsActive).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build insert")
	}
	return r.pool.QueryRow(ctx, sql, args...).Scan(&w.CreatedAt, &w.UpdatedAt)
}
