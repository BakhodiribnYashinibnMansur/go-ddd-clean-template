package postgres

import (
	"context"

	"gct/internal/ratelimit/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var readColumns = []string{
	"id", "name", "path_pattern", "limit_count", "window_seconds",
	"is_active", "created_at", "updated_at",
}

// RateLimitReadRepo implements domain.RateLimitReadRepository for the CQRS read side.
type RateLimitReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewRateLimitReadRepo creates a new RateLimitReadRepo.
func NewRateLimitReadRepo(pool *pgxpool.Pool) *RateLimitReadRepo {
	return &RateLimitReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// FindByID returns a single RateLimitView by ID.
func (r *RateLimitReadRepo) FindByID(ctx context.Context, id uuid.UUID) (result *domain.RateLimitView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RateLimitReadRepo.FindByID")
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
	return scanRateLimitView(row)
}

// List returns a paginated list of RateLimitView with optional filters.
func (r *RateLimitReadRepo) List(ctx context.Context, filter domain.RateLimitFilter) (views []*domain.RateLimitView, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "RateLimitReadRepo.List")
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
		Select(readColumns...).
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
		v, err := scanRateLimitViewFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		views = append(views, v)
	}

	return views, total, nil
}

func scanRateLimitView(row pgx.Row) (*domain.RateLimitView, error) {
	var v domain.RateLimitView
	err := row.Scan(&v.ID, &v.Name, &v.Rule, &v.RequestsPerWindow, &v.WindowDuration, &v.Enabled, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}
	return &v, nil
}

func scanRateLimitViewFromRows(rows pgx.Rows) (*domain.RateLimitView, error) {
	var v domain.RateLimitView
	err := rows.Scan(&v.ID, &v.Name, &v.Rule, &v.RequestsPerWindow, &v.WindowDuration, &v.Enabled, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
