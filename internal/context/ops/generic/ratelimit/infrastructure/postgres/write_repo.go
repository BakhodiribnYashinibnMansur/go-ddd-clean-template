package postgres

import (
	"context"
	"time"

	ratelimitentity "gct/internal/context/ops/generic/ratelimit/domain/entity"
	ratelimitrepo "gct/internal/context/ops/generic/ratelimit/domain/repository"
	"gct/internal/kernel/consts"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableRateLimits

var writeColumns = []string{
	"id", "name", "path_pattern", "limit_count", "window_seconds",
	"is_active", "created_at", "updated_at",
}

// RateLimitWriteRepo implements ratelimitrepo.RateLimitRepository using PostgreSQL.
type RateLimitWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewRateLimitWriteRepo creates a new RateLimitWriteRepo.
func NewRateLimitWriteRepo(pool *pgxpool.Pool) *RateLimitWriteRepo {
	return &RateLimitWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new RateLimit aggregate into the database.
func (r *RateLimitWriteRepo) Save(ctx context.Context, rl *ratelimitentity.RateLimit) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RateLimitWriteRepo.Save")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			rl.ID(), rl.Name(), rl.Rule(), rl.RequestsPerWindow(),
			rl.WindowDuration(), rl.Enabled(),
			rl.CreatedAt(), rl.UpdatedAt(),
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = pgxutil.QuerierFromContext(ctx, r.pool).Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// FindByID retrieves a RateLimit aggregate by its ID.
func (r *RateLimitWriteRepo) FindByID(ctx context.Context, id ratelimitentity.RateLimitID) (result *ratelimitentity.RateLimit, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RateLimitWriteRepo.FindByID")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select(writeColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id.UUID()}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanRateLimit(row)
}

// Update updates an existing RateLimit aggregate in the database.
func (r *RateLimitWriteRepo) Update(ctx context.Context, rl *ratelimitentity.RateLimit) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RateLimitWriteRepo.Update")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Update(tableName).
		Set("name", rl.Name()).
		Set("path_pattern", rl.Rule()).
		Set("limit_count", rl.RequestsPerWindow()).
		Set("window_seconds", rl.WindowDuration()).
		Set("is_active", rl.Enabled()).
		Set("updated_at", rl.UpdatedAt()).
		Where(squirrel.Eq{"id": rl.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = pgxutil.QuerierFromContext(ctx, r.pool).Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// Delete removes a RateLimit by its ID.
func (r *RateLimitWriteRepo) Delete(ctx context.Context, id ratelimitentity.RateLimitID) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RateLimitWriteRepo.Delete")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Delete(tableName).
		Where(squirrel.Eq{"id": id.UUID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = pgxutil.QuerierFromContext(ctx, r.pool).Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// List retrieves a paginated list of RateLimit aggregates with optional filters.
func (r *RateLimitWriteRepo) List(ctx context.Context, filter ratelimitrepo.RateLimitFilter) (results []*ratelimitentity.RateLimit, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RateLimitWriteRepo.List")
	defer func() { end(err) }()

	conds := squirrel.And{}
	conds = applyFilters(conds, filter)

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

	sqlStr, args, err := qb.ToSql()
	if err != nil {
		return nil, 0, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	rows, err := r.pool.Query(ctx, sqlStr, args...)
	if err != nil {
		return nil, 0, apperrors.HandlePgError(err, tableName, nil)
	}
	defer rows.Close()

	for rows.Next() {
		rl, err := scanRateLimitFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		results = append(results, rl)
	}

	return results, total, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func applyFilters(conds squirrel.And, filter ratelimitrepo.RateLimitFilter) squirrel.And {
	if filter.Name != nil {
		conds = append(conds, squirrel.Eq{"name": *filter.Name})
	}
	if filter.Enabled != nil {
		conds = append(conds, squirrel.Eq{"is_active": *filter.Enabled})
	}
	return conds
}

func scanRateLimit(row pgx.Row) (*ratelimitentity.RateLimit, error) {
	var (
		id                uuid.UUID
		name              string
		rule              string
		requestsPerWindow int
		windowDuration    int
		enabled           bool
		createdAt         time.Time
		updatedAt         time.Time
	)

	err := row.Scan(&id, &name, &rule, &requestsPerWindow, &windowDuration, &enabled, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return ratelimitentity.ReconstructRateLimit(id, createdAt, updatedAt, name, rule, requestsPerWindow, windowDuration, enabled), nil
}

func scanRateLimitFromRows(rows pgx.Rows) (*ratelimitentity.RateLimit, error) {
	var (
		id                uuid.UUID
		name              string
		rule              string
		requestsPerWindow int
		windowDuration    int
		enabled           bool
		createdAt         time.Time
		updatedAt         time.Time
	)

	err := rows.Scan(&id, &name, &rule, &requestsPerWindow, &windowDuration, &enabled, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return ratelimitentity.ReconstructRateLimit(id, createdAt, updatedAt, name, rule, requestsPerWindow, windowDuration, enabled), nil
}
