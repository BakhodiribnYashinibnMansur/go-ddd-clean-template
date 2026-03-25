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
	"id", "task_name", "status", "payload", "result",
	"attempts", "max_attempts", "scheduled_at", "started_at",
	"completed_at", "error", "created_at", "updated_at",
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
	resultJSON, _ := json.Marshal(j.Result())

	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			j.ID(),
			j.TaskName(),
			j.Status(),
			payloadJSON,
			resultJSON,
			j.Attempts(),
			j.MaxAttempts(),
			j.ScheduledAt(),
			j.StartedAt(),
			j.CompletedAt(),
			j.Error(),
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
	resultJSON, _ := json.Marshal(j.Result())

	sql, args, err := r.builder.
		Update(tableName).
		Set("task_name", j.TaskName()).
		Set("status", j.Status()).
		Set("payload", payloadJSON).
		Set("result", resultJSON).
		Set("attempts", j.Attempts()).
		Set("max_attempts", j.MaxAttempts()).
		Set("scheduled_at", j.ScheduledAt()).
		Set("started_at", j.StartedAt()).
		Set("completed_at", j.CompletedAt()).
		Set("error", j.Error()).
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
		id          uuid.UUID
		taskName    string
		status      string
		payloadJSON []byte
		resultJSON  []byte
		attempts    int
		maxAttempts int
		scheduledAt *time.Time
		startedAt   *time.Time
		completedAt *time.Time
		errorMsg    *string
		createdAt   time.Time
		updatedAt   time.Time
	)

	err := row.Scan(
		&id, &taskName, &status, &payloadJSON, &resultJSON,
		&attempts, &maxAttempts, &scheduledAt, &startedAt,
		&completedAt, &errorMsg, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	var payload map[string]any
	_ = json.Unmarshal(payloadJSON, &payload)

	var result map[string]any
	_ = json.Unmarshal(resultJSON, &result)

	return domain.ReconstructJob(
		id, createdAt, updatedAt,
		taskName, status, payload, result,
		attempts, maxAttempts,
		scheduledAt, startedAt, completedAt,
		errorMsg,
	), nil
}
