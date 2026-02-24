package job

import (
	"context"
	"encoding/json"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, j *domain.Job) error {
	payloadJSON, err := json.Marshal(j.Payload)
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "marshal payload")
	}
	sql, args, err := r.builder.
		Insert(table).
		Columns("id", "name", "type", "cron_schedule", "payload", "is_active", "status").
		Values(j.ID, j.Name, j.Type, j.CronSchedule, payloadJSON, j.IsActive, j.Status).
		Suffix("RETURNING created_at, updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build insert")
	}
	return r.pool.QueryRow(ctx, sql, args...).Scan(&j.CreatedAt, &j.UpdatedAt)
}
