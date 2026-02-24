package webhook

import (
	"context"
	"encoding/json"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) List(ctx context.Context, filter domain.WebhookFilter) ([]domain.Webhook, int64, error) {
	q := r.builder.
		Select("id", "name", "url", "secret", "events", "headers", "is_active", "created_at", "updated_at").
		From(table).
		Where(squirrel.Eq{"deleted_at": nil})

	if filter.Search != "" {
		q = q.Where(squirrel.ILike{"name": "%" + filter.Search + "%"})
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

	var items []domain.Webhook
	for rows.Next() {
		var w domain.Webhook
		var eventsRaw, headersRaw []byte
		if err := rows.Scan(&w.ID, &w.Name, &w.URL, &w.Secret, &eventsRaw, &headersRaw, &w.IsActive, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, 0, apperrors.HandlePgError(err, table, nil)
		}
		if err := json.Unmarshal(eventsRaw, &w.Events); err != nil {
			w.Events = []string{}
		}
		if err := json.Unmarshal(headersRaw, &w.Headers); err != nil {
			w.Headers = map[string]any{}
		}
		items = append(items, w)
	}
	if items == nil {
		items = []domain.Webhook{}
	}
	return items, total, nil
}
