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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableSystemError

var writeColumns = []string{
	"id", "code", "message", "stack_trace", "metadata",
	"severity", "service_name", "request_id", "user_id",
	"ip_address", "path", "method",
	"is_resolved", "resolved_at", "resolved_by", "created_at",
}

// SystemErrorWriteRepo implements domain.SystemErrorRepository using PostgreSQL.
type SystemErrorWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewSystemErrorWriteRepo creates a new SystemErrorWriteRepo.
func NewSystemErrorWriteRepo(pool *pgxpool.Pool) *SystemErrorWriteRepo {
	return &SystemErrorWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new SystemError aggregate into the database.
func (r *SystemErrorWriteRepo) Save(ctx context.Context, se *domain.SystemError) error {
	metaJSON, err := json.Marshal(se.Metadata())
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToMarshalJSON)
	}

	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			se.ID(),
			se.Code(),
			se.Message(),
			se.StackTrace(),
			metaJSON,
			se.Severity(),
			se.ServiceName(),
			se.RequestID(),
			se.UserID(),
			se.IPAddress(),
			se.Path(),
			se.Method(),
			se.IsResolved(),
			se.ResolvedAt(),
			se.ResolvedBy(),
			se.CreatedAt(),
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

// FindByID retrieves a SystemError aggregate by ID.
func (r *SystemErrorWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.SystemError, error) {
	sql, args, err := r.builder.
		Select(writeColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	se, err := scanSystemError(row)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, map[string]any{"id": id})
	}
	return se, nil
}

// Update updates the SystemError aggregate in the database.
func (r *SystemErrorWriteRepo) Update(ctx context.Context, se *domain.SystemError) error {
	metaJSON, err := json.Marshal(se.Metadata())
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToMarshalJSON)
	}

	sql, args, err := r.builder.
		Update(tableName).
		Set("code", se.Code()).
		Set("message", se.Message()).
		Set("stack_trace", se.StackTrace()).
		Set("metadata", metaJSON).
		Set("severity", se.Severity()).
		Set("service_name", se.ServiceName()).
		Set("request_id", se.RequestID()).
		Set("user_id", se.UserID()).
		Set("ip_address", se.IPAddress()).
		Set("path", se.Path()).
		Set("method", se.Method()).
		Set("is_resolved", se.IsResolved()).
		Set("resolved_at", se.ResolvedAt()).
		Set("resolved_by", se.ResolvedBy()).
		Where(squirrel.Eq{"id": se.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// List retrieves a paginated list of SystemError aggregates with optional filters.
func (r *SystemErrorWriteRepo) List(ctx context.Context, filter domain.SystemErrorFilter) ([]*domain.SystemError, int64, error) {
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
		Select(writeColumns...).
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

	var results []*domain.SystemError
	for rows.Next() {
		se, err := scanSystemErrorFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		results = append(results, se)
	}

	return results, total, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func applyFilters(conds squirrel.And, filter domain.SystemErrorFilter) squirrel.And {
	if filter.Code != nil {
		conds = append(conds, squirrel.Eq{"code": *filter.Code})
	}
	if filter.Severity != nil {
		conds = append(conds, squirrel.Eq{"severity": *filter.Severity})
	}
	if filter.IsResolved != nil {
		conds = append(conds, squirrel.Eq{"is_resolved": *filter.IsResolved})
	}
	if filter.FromDate != nil {
		conds = append(conds, squirrel.GtOrEq{"created_at": *filter.FromDate})
	}
	if filter.ToDate != nil {
		conds = append(conds, squirrel.LtOrEq{"created_at": *filter.ToDate})
	}
	if filter.RequestID != nil {
		conds = append(conds, squirrel.Eq{"request_id": *filter.RequestID})
	}
	if filter.UserID != nil {
		conds = append(conds, squirrel.Eq{"user_id": *filter.UserID})
	}
	return conds
}

func scanSystemError(row pgx.Row) (*domain.SystemError, error) {
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

	err := row.Scan(
		&id, &code, &message, &stackTrace, &metaJSON,
		&severity, &serviceName, &requestID, &userID,
		&ipAddress, &path, &method,
		&isResolved, &resolvedAt, &resolvedBy, &createdAt,
	)
	if err != nil {
		return nil, err
	}

	return reconstructFromRow(
		id, createdAt, code, message, stackTrace, metaJSON,
		severity, serviceName, requestID, userID,
		ipAddress, path, method,
		isResolved, resolvedAt, resolvedBy,
	), nil
}

func scanSystemErrorFromRows(rows pgx.Rows) (*domain.SystemError, error) {
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

	err := rows.Scan(
		&id, &code, &message, &stackTrace, &metaJSON,
		&severity, &serviceName, &requestID, &userID,
		&ipAddress, &path, &method,
		&isResolved, &resolvedAt, &resolvedBy, &createdAt,
	)
	if err != nil {
		return nil, err
	}

	return reconstructFromRow(
		id, createdAt, code, message, stackTrace, metaJSON,
		severity, serviceName, requestID, userID,
		ipAddress, path, method,
		isResolved, resolvedAt, resolvedBy,
	), nil
}

func reconstructFromRow(
	id uuid.UUID,
	createdAt time.Time,
	code, message string,
	stackTrace *string,
	metaJSON []byte,
	severity string,
	serviceName *string,
	requestID *uuid.UUID,
	userID *uuid.UUID,
	ipAddress *string,
	path *string,
	method *string,
	isResolved bool,
	resolvedAt *time.Time,
	resolvedBy *uuid.UUID,
) *domain.SystemError {
	var metadata map[string]any
	if len(metaJSON) > 0 {
		_ = json.Unmarshal(metaJSON, &metadata)
	}

	return domain.ReconstructSystemError(
		id, createdAt,
		code, message, stackTrace, metadata,
		severity, serviceName, requestID, userID,
		ipAddress, path, method,
		isResolved, resolvedAt, resolvedBy,
	)
}
