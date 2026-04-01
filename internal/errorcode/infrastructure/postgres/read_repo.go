package postgres

import (
	"context"
	"time"

	"gct/internal/errorcode/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var readColumns = []string{
	"id", "code", "message", "http_status", "category", "severity",
	"retryable", "retry_after", "suggestion", "created_at", "updated_at",
}

// ErrorCodeReadRepo implements domain.ErrorCodeReadRepository for the CQRS read side.
type ErrorCodeReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewErrorCodeReadRepo creates a new ErrorCodeReadRepo.
func NewErrorCodeReadRepo(pool *pgxpool.Pool) *ErrorCodeReadRepo {
	return &ErrorCodeReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// FindByID returns a single ErrorCodeView by its ID.
func (r *ErrorCodeReadRepo) FindByID(ctx context.Context, id uuid.UUID) (result *domain.ErrorCodeView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "ErrorCodeReadRepo.FindByID")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select(readColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanErrorCodeView(row)
}

// List returns a paginated list of ErrorCodeView with optional filters.
func (r *ErrorCodeReadRepo) List(ctx context.Context, filter domain.ErrorCodeFilter) (items []*domain.ErrorCodeView, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "ErrorCodeReadRepo.List")
	defer func() { end(err) }()

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

	var views []*domain.ErrorCodeView
	for rows.Next() {
		v, err := scanErrorCodeViewFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		views = append(views, v)
	}

	return views, total, nil
}

func applyFilters(conds squirrel.And, filter domain.ErrorCodeFilter) squirrel.And {
	if filter.Code != nil {
		conds = append(conds, squirrel.Eq{"code": *filter.Code})
	}
	if filter.Category != nil {
		conds = append(conds, squirrel.Eq{"category": *filter.Category})
	}
	if filter.Severity != nil {
		conds = append(conds, squirrel.Eq{"severity": *filter.Severity})
	}
	return conds
}

func scanErrorCodeView(row pgx.Row) (*domain.ErrorCodeView, error) {
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

	return &domain.ErrorCodeView{
		ID:         id,
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Category:   category,
		Severity:   severity,
		Retryable:  retryable,
		RetryAfter: retryAfter,
		Suggestion: suggestion,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

func scanErrorCodeViewFromRows(rows pgx.Rows) (*domain.ErrorCodeView, error) {
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

	err := rows.Scan(
		&id, &code, &message, &httpStatus, &category, &severity,
		&retryable, &retryAfter, &suggestion, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &domain.ErrorCodeView{
		ID:         id,
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Category:   category,
		Severity:   severity,
		Retryable:  retryable,
		RetryAfter: retryAfter,
		Suggestion: suggestion,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}
