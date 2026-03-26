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

var readColumns = []string{
	"id", "name", "type", "cron_schedule", "payload",
	"is_active", "status", "last_run_at", "next_run_at",
	"created_at", "updated_at",
}

// JobReadRepo implements domain.JobReadRepository for the CQRS read side.
type JobReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewJobReadRepo creates a new JobReadRepo.
func NewJobReadRepo(pool *pgxpool.Pool) *JobReadRepo {
	return &JobReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// FindByID returns a single JobView by its ID.
func (r *JobReadRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.JobView, error) {
	sql, args, err := r.builder.
		Select(readColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanJobView(row)
}

// List returns a paginated list of JobView with optional filters.
func (r *JobReadRepo) List(ctx context.Context, filter domain.JobFilter) ([]*domain.JobView, int64, error) {
	conds := squirrel.And{}
	conds = applyFilters(conds, filter)

	// Count total.
	countQB := r.builder.Select("COUNT(*)").From(tableName)
	if len(conds) > 0 {
		countQB = countQB.Where(conds)
	}
	countSQL, countArgs, err := countQB.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var total int64
	if err = r.pool.QueryRow(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}

	// Fetch page.
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	qb := r.builder.
		Select(readColumns...).
		From(tableName).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(filter.Offset))

	if len(conds) > 0 {
		qb = qb.Where(conds)
	}

	sql, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}
	defer rows.Close()

	var views []*domain.JobView
	for rows.Next() {
		v, err := scanJobViewFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		views = append(views, v)
	}

	return views, total, nil
}

func applyFilters(conds squirrel.And, filter domain.JobFilter) squirrel.And {
	if filter.TaskName != nil {
		conds = append(conds, squirrel.Eq{"name": *filter.TaskName})
	}
	if filter.Status != nil {
		conds = append(conds, squirrel.Eq{"status": *filter.Status})
	}
	return conds
}

func scanJobView(row pgx.Row) (*domain.JobView, error) {
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

	return &domain.JobView{
		ID:          id,
		TaskName:    name,
		Status:      status,
		Payload:     payload,
		Result:      nil,
		Attempts:    0,
		MaxAttempts: 0,
		ScheduledAt: nextRunAt,
		StartedAt:   lastRunAt,
		CompletedAt: nil,
		Error:       nil,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}

func scanJobViewFromRows(rows pgx.Rows) (*domain.JobView, error) {
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

	err := rows.Scan(
		&id, &name, &jobType, &cronSchedule, &payloadJSON,
		&isActive, &status, &lastRunAt, &nextRunAt,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	_ = jobType
	_ = cronSchedule
	_ = isActive

	var payload map[string]any
	_ = json.Unmarshal(payloadJSON, &payload)

	return &domain.JobView{
		ID:          id,
		TaskName:    name,
		Status:      status,
		Payload:     payload,
		Result:      nil,
		Attempts:    0,
		MaxAttempts: 0,
		ScheduledAt: nextRunAt,
		StartedAt:   lastRunAt,
		CompletedAt: nil,
		Error:       nil,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}
