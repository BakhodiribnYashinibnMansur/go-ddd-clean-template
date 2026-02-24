package webhook

import (
	"context"
	"encoding/json"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Webhook, error) {
	sql, args, err := r.builder.
		Select("id", "name", "url", "secret", "events", "headers", "is_active", "created_at", "updated_at", "deleted_at").
		From(table).
		Where(squirrel.Eq{"id": id, "deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build select")
	}
	var w domain.Webhook
	var eventsRaw, headersRaw []byte
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&w.ID, &w.Name, &w.URL, &w.Secret, &eventsRaw, &headersRaw,
		&w.IsActive, &w.CreatedAt, &w.UpdatedAt, &w.DeletedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, table, nil)
	}
	if err := json.Unmarshal(eventsRaw, &w.Events); err != nil {
		w.Events = []string{}
	}
	if err := json.Unmarshal(headersRaw, &w.Headers); err != nil {
		w.Headers = map[string]any{}
	}
	return &w, nil
}
