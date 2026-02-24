package job

import (
	"context"
	"encoding/json"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) List(ctx context.Context, filter domain.JobFilter) ([]domain.Job, int64, error) {
	q := r.builder.
		Select("id", "name", "type", "cron_schedule", "payload", "is_active", "status", "last_run_at", "next_run_at", "created_at", "updated_at").
		From(table)

	if filter.Search != "" {
		q = q.Where(squirrel.ILike{"name": "%" + filter.Search + "%"})
	}
	if filter.Status != "" {
		q = q.Where(squirrel.Eq{"status": filter.Status})
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

	var items []domain.Job
	for rows.Next() {
		var j domain.Job
		var payloadRaw []byte
		if err := rows.Scan(&j.ID, &j.Name, &j.Type, &j.CronSchedule, &payloadRaw, &j.IsActive, &j.Status, &j.LastRunAt, &j.NextRunAt, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, 0, apperrors.HandlePgError(err, table, nil)
		}
		if err := json.Unmarshal(payloadRaw, &j.Payload); err != nil {
			j.Payload = map[string]any{}
		}
		items = append(items, j)
	}
	if items == nil {
		items = []domain.Job{}
	}
	return items, total, nil
}
