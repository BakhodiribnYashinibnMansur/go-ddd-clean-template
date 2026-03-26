package postgres

import (
	"context"
	"encoding/json"
	"time"

	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/systemerror/domain"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var readColumns = []string{
	"id", "code", "message", "stack_trace", "metadata",
	"severity", "service_name", "request_id", "user_id",
	"ip_address::text", "path", "method",
	"is_resolved", "resolved_at", "resolved_by", "created_at",
}

// SystemErrorReadRepo implements domain.SystemErrorReadRepository for the CQRS read side.
type SystemErrorReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewSystemErrorReadRepo creates a new SystemErrorReadRepo.
func NewSystemErrorReadRepo(pool *pgxpool.Pool) *SystemErrorReadRepo {
	return &SystemErrorReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// FindByID returns a SystemErrorView for the given ID.
func (r *SystemErrorReadRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.SystemErrorView, error) {
	sql, args, err := r.builder.
		Select(readColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanView(row)
}

// List returns a paginated list of SystemErrorView with optional filters.
func (r *SystemErrorReadRepo) List(ctx context.Context, filter domain.SystemErrorFilter) ([]*domain.SystemErrorView, int64, error) {
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

	var views []*domain.SystemErrorView
	for rows.Next() {
		v, err := scanViewFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		views = append(views, v)
	}

	return views, total, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

type rowScanner interface {
	Scan(dest ...any) error
}

func scanViewFields(s rowScanner) (*domain.SystemErrorView, error) {
	var (
		id          uuid.UUID
		code        string
		message     string
		stackTrace  *string
		metaJSON    []byte
		severity    string
		serviceName *string
		requestID   *uuid.UUID
		userID      *uuid.UUID
		ipAddress   *string
		path        *string
		method      *string
		isResolved  bool
		resolvedAt  *time.Time
		resolvedBy  *uuid.UUID
		createdAt   time.Time
	)

	err := s.Scan(
		&id, &code, &message, &stackTrace, &metaJSON,
		&severity, &serviceName, &requestID, &userID,
		&ipAddress, &path, &method,
		&isResolved, &resolvedAt, &resolvedBy, &createdAt,
	)
	if err != nil {
		return nil, err
	}

	var metadata map[string]any
	if len(metaJSON) > 0 {
		_ = json.Unmarshal(metaJSON, &metadata)
	}

	return &domain.SystemErrorView{
		ID:          id,
		Code:        code,
		Message:     message,
		StackTrace:  stackTrace,
		Metadata:    metadata,
		Severity:    severity,
		ServiceName: serviceName,
		RequestID:   requestID,
		UserID:      userID,
		IPAddress:   ipAddress,
		Path:        path,
		Method:      method,
		IsResolved:  isResolved,
		ResolvedAt:  resolvedAt,
		ResolvedBy:  resolvedBy,
		CreatedAt:   createdAt,
	}, nil
}

func scanView(row rowScanner) (*domain.SystemErrorView, error) {
	v, err := scanViewFields(row)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}
	return v, nil
}

func scanViewFromRows(rows rowScanner) (*domain.SystemErrorView, error) {
	return scanViewFields(rows)
}
