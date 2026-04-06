package postgres

import (
	"context"
	"time"

	metricentity "gct/internal/context/ops/generic/metric/domain/entity"
	metricrepo "gct/internal/context/ops/generic/metric/domain/repository"
	"gct/internal/kernel/consts"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableFunctionMetric

var writeColumns = []string{
	"id", "name", "latency_ms", "is_panic", "panic_error", "created_at",
}

// MetricWriteRepo implements domain.MetricRepository using PostgreSQL.
type MetricWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewMetricWriteRepo creates a new MetricWriteRepo.
func NewMetricWriteRepo(pool *pgxpool.Pool) *MetricWriteRepo {
	return &MetricWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new FunctionMetric aggregate into the database.
func (r *MetricWriteRepo) Save(ctx context.Context, fm *metricentity.FunctionMetric) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "MetricWriteRepo.Save")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			fm.ID(),
			fm.Name(),
			fm.LatencyMs(),
			fm.IsPanic(),
			fm.PanicError(),
			fm.CreatedAt(),
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

// List retrieves a paginated list of FunctionMetric aggregates with optional filters.
func (r *MetricWriteRepo) List(ctx context.Context, filter metricrepo.MetricFilter) (items []*metricentity.FunctionMetric, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "MetricWriteRepo.List")
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

	var results []*metricentity.FunctionMetric
	for rows.Next() {
		fm, err := scanMetricFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		results = append(results, fm)
	}

	return results, total, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func applyFilters(conds squirrel.And, filter metricrepo.MetricFilter) squirrel.And {
	if filter.Name != nil {
		conds = append(conds, squirrel.Eq{"name": *filter.Name})
	}
	if filter.IsPanic != nil {
		conds = append(conds, squirrel.Eq{"is_panic": *filter.IsPanic})
	}
	if filter.FromDate != nil {
		conds = append(conds, squirrel.GtOrEq{"created_at": *filter.FromDate})
	}
	if filter.ToDate != nil {
		conds = append(conds, squirrel.LtOrEq{"created_at": *filter.ToDate})
	}
	return conds
}

func scanMetricFromRows(rows pgx.Rows) (*metricentity.FunctionMetric, error) {
	var (
		id         uuid.UUID
		name       string
		latencyMs  float64
		isPanic    bool
		panicError *string
		createdAt  time.Time
	)

	err := rows.Scan(&id, &name, &latencyMs, &isPanic, &panicError, &createdAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return metricentity.ReconstructFunctionMetric(id, createdAt, name, latencyMs, isPanic, panicError), nil
}
