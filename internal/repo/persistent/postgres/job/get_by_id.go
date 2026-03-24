package job

import (
	"context"
	"encoding/json"

	"gct/internal/domain"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *Repo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	sql, args, err := r.builder.
		Select("id", "name", "type", "cron_schedule", "payload", "is_active", "status", "last_run_at", "next_run_at", "created_at", "updated_at").
		From(table).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "build select")
	}
	var j domain.Job
	var payloadRaw []byte
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&j.ID, &j.Name, &j.Type, &j.CronSchedule, &payloadRaw,
		&j.IsActive, &j.Status, &j.LastRunAt, &j.NextRunAt, &j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, table, nil)
	}
	if err := json.Unmarshal(payloadRaw, &j.Payload); err != nil {
		j.Payload = map[string]any{}
	}
	return &j, nil
}
