package postgres

import (
	"context"
	"time"

	"gct/internal/errorcode/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = "error_codes"

var writeColumns = []string{
	"id", "code", "message", "http_status", "category", "severity",
	"retryable", "retry_after", "suggestion", "created_at", "updated_at",
}

// ErrorCodeWriteRepo implements domain.ErrorCodeRepository using PostgreSQL.
type ErrorCodeWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewErrorCodeWriteRepo creates a new ErrorCodeWriteRepo.
func NewErrorCodeWriteRepo(pool *pgxpool.Pool) *ErrorCodeWriteRepo {
	return &ErrorCodeWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new ErrorCode aggregate into the database.
func (r *ErrorCodeWriteRepo) Save(ctx context.Context, ec *domain.ErrorCode) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			ec.ID(),
			ec.Code(),
			ec.Message(),
			ec.HTTPStatus(),
			ec.Category(),
			ec.Severity(),
			ec.Retryable(),
			ec.RetryAfter(),
			ec.Suggestion(),
			ec.CreatedAt(),
			ec.UpdatedAt(),
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

// Update updates an existing ErrorCode aggregate in the database.
func (r *ErrorCodeWriteRepo) Update(ctx context.Context, ec *domain.ErrorCode) error {
	sql, args, err := r.builder.
		Update(tableName).
		Set("message", ec.Message()).
		Set("http_status", ec.HTTPStatus()).
		Set("category", ec.Category()).
		Set("severity", ec.Severity()).
		Set("retryable", ec.Retryable()).
		Set("retry_after", ec.RetryAfter()).
		Set("suggestion", ec.Suggestion()).
		Set("updated_at", ec.UpdatedAt()).
		Where(squirrel.Eq{"id": ec.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// FindByID retrieves an ErrorCode aggregate by its ID.
func (r *ErrorCodeWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.ErrorCode, error) {
	sql, args, err := r.builder.
		Select(writeColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanErrorCode(row)
}

// Delete removes an error code by its ID.
func (r *ErrorCodeWriteRepo) Delete(ctx context.Context, id uuid.UUID) error {
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

func scanErrorCode(row pgx.Row) (*domain.ErrorCode, error) {
	var (
		id         uuid.UUID
		code       string
		message    string
		httpStatus int
		category   string
		severity   string
		retryable  bool
		retryAfter int
		suggestion string
		createdAt  time.Time
		updatedAt  time.Time
	)

	err := row.Scan(
		&id, &code, &message, &httpStatus, &category, &severity,
		&retryable, &retryAfter, &suggestion, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return domain.ReconstructErrorCode(
		id, createdAt, updatedAt,
		code, message, httpStatus,
		category, severity, retryable,
		retryAfter, suggestion,
	), nil
}
