package postgres

import (
	"context"
	"encoding/json"
	"time"

	"gct/internal/job/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableJobs

var writeColumns = []string{
	"id", "name", "type", "cron_schedule", "payload",
	"is_active", "status", "last_run_at", "next_run_at",
	"created_at", "updated_at",
}

// JobWriteRepo implements domain.JobRepository using PostgreSQL.
type JobWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewJobWriteRepo creates a new JobWriteRepo.
func NewJobWriteRepo(pool *pgxpool.Pool) *JobWriteRepo {
	return &JobWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new Job aggregate into the database.
func (r *JobWriteRepo) Save(ctx context.Context, j *domain.Job) error {
	payloadJSON, _ := json.Marshal(j.Payload())

	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			j.ID(),
			j.TaskName(),
			"default",
			"",
			payloadJSON,
			true,
			j.Status(),
			j.StartedAt(),
			j.ScheduledAt(),
			j.CreatedAt(),
			j.UpdatedAt(),
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// Update updates an existing Job aggregate in the database.
func (r *JobWriteRepo) Update(ctx context.Context, j *domain.Job) error {
	payloadJSON, _ := json.Marshal(j.Payload())

	sql, args, err := r.builder.
		Update(tableName).
		Set("name", j.TaskName()).
		Set("status", j.Status()).
		Set("payload", payloadJSON).
		Set("last_run_at", j.StartedAt()).
		Set("next_run_at", j.ScheduledAt()).
		Set("updated_at", j.UpdatedAt()).
		Where(squirrel.Eq{"id": j.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// FindByID retrieves a Job aggregate by its ID.
func (r *JobWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	sql, args, err := r.builder.
		Select(writeColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanJob(row)
}

// Delete removes a job by its ID.
func (r *JobWriteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := r.builder.
		Delete(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

func scanJob(row pgx.Row) (*domain.Job, error) {
	var (
		id           uuid.UUID
		name         string
		jobType      string
		cronSchedule string
		payloadJSON  []byte
		isActive     bool
		status       string
		lastRunAt    *time.Time
		nextRunAt    *time.Time
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := row.Scan(
		&id, &name, &jobType, &cronSchedule, &payloadJSON,
		&isActive, &status, &lastRunAt, &nextRunAt,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	_ = jobType
	_ = cronSchedule
	_ = isActive

	var payload map[string]any
	_ = json.Unmarshal(payloadJSON, &payload)

	return domain.ReconstructJob(
		id, createdAt, updatedAt,
		name, status, payload, nil,
		0, 0,
		nextRunAt, lastRunAt, nil,
		nil,
	), nil
}
