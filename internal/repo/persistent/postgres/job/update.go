package job

import (
	"context"
	"encoding/json"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Update(ctx context.Context, j *domain.Job) error {
	payloadJSON, err := json.Marshal(j.Payload)
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "marshal payload")
	}
	sql, args, err := r.builder.
		Update(table).
		Set("name", j.Name).
		Set("type", j.Type).
		Set("cron_schedule", j.CronSchedule).
		Set("payload", payloadJSON).
		Set("is_active", j.IsActive).
		Set("status", j.Status).
		Set("last_run_at", j.LastRunAt).
		Set("next_run_at", j.NextRunAt).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": j.ID}).
		Suffix("RETURNING updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build update")
	}
	return r.pool.QueryRow(ctx, sql, args...).Scan(&j.UpdatedAt)
}
