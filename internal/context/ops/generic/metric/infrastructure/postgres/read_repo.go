package postgres

import (
	"context"
	"time"

	"gct/internal/context/ops/generic/metric/domain"
	"gct/internal/kernel/consts"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var readColumns = []string{
	"id", "name", "latency_ms", "is_panic", "panic_error", "created_at",
}

// MetricReadRepo implements domain.MetricReadRepository for the CQRS read side.
type MetricReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewMetricReadRepo creates a new MetricReadRepo.
func NewMetricReadRepo(pool *pgxpool.Pool) *MetricReadRepo {
	return &MetricReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// List returns a paginated list of MetricView with optional filters.
func (r *MetricReadRepo) List(ctx context.Context, filter domain.MetricFilter) (items []*domain.MetricView, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "MetricReadRepo.List")
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

	var views []*domain.MetricView
	for rows.Next() {
		v, err := scanMetricView(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		views = append(views, v)
	}

	return views, total, nil
}

func scanMetricView(rows pgx.Rows) (*domain.MetricView, error) {
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

	return &domain.MetricView{
		ID:         domain.MetricID(id),
		Name:       name,
		LatencyMs:  latencyMs,
		IsPanic:    isPanic,
		PanicError: panicError,
		CreatedAt:  createdAt,
	}, nil
}
