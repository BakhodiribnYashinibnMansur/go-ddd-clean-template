package postgres

import (
	"context"

	"gct/internal/iprule/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var readColumns = []string{
	"id", "ip_address", "type", "reason", "is_active", "created_at", "updated_at",
}

// IPRuleReadRepo implements domain.IPRuleReadRepository for the CQRS read side.
type IPRuleReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewIPRuleReadRepo creates a new IPRuleReadRepo.
func NewIPRuleReadRepo(pool *pgxpool.Pool) *IPRuleReadRepo {
	return &IPRuleReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// FindByID returns a single IPRuleView by ID.
func (r *IPRuleReadRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.IPRuleView, error) {
	sql, args, err := r.builder.
		Select(readColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanIPRuleView(row)
}

// List returns a paginated list of IPRuleView with optional filters.
func (r *IPRuleReadRepo) List(ctx context.Context, filter domain.IPRuleFilter) ([]*domain.IPRuleView, int64, error) {
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

	var total int64
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

	var views []*domain.IPRuleView
	for rows.Next() {
		v, err := scanIPRuleViewFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		views = append(views, v)
	}

	return views, total, nil
}

func scanIPRuleView(row pgx.Row) (*domain.IPRuleView, error) {
	var (
		v        domain.IPRuleView
		isActive bool
	)
	err := row.Scan(&v.ID, &v.IPAddress, &v.Action, &v.Reason, &isActive, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}
	_ = isActive
	v.ExpiresAt = nil
	return &v, nil
}

func scanIPRuleViewFromRows(rows pgx.Rows) (*domain.IPRuleView, error) {
	var (
		v        domain.IPRuleView
		isActive bool
	)
	err := rows.Scan(&v.ID, &v.IPAddress, &v.Action, &v.Reason, &isActive, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, err
	}
	_ = isActive
	v.ExpiresAt = nil
	return &v, nil
}
