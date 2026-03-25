package postgres

import (
	"context"

	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/sitesetting/domain"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var readColumns = []string{
	"id", "key", "value", "type", "description", "created_at", "updated_at",
}

// SiteSettingReadRepo implements domain.SiteSettingReadRepository for the CQRS read side.
type SiteSettingReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewSiteSettingReadRepo creates a new SiteSettingReadRepo.
func NewSiteSettingReadRepo(pool *pgxpool.Pool) *SiteSettingReadRepo {
	return &SiteSettingReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// FindByID returns a single SiteSettingView by ID.
func (r *SiteSettingReadRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.SiteSettingView, error) {
	sql, args, err := r.builder.
		Select(readColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanSiteSettingView(row)
}

// List returns a paginated list of SiteSettingView with optional filters.
func (r *SiteSettingReadRepo) List(ctx context.Context, filter domain.SiteSettingFilter) ([]*domain.SiteSettingView, int64, error) {
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

	var views []*domain.SiteSettingView
	for rows.Next() {
		v, err := scanSiteSettingViewFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		views = append(views, v)
	}

	return views, total, nil
}

func scanSiteSettingView(row pgx.Row) (*domain.SiteSettingView, error) {
	var v domain.SiteSettingView
	err := row.Scan(&v.ID, &v.Key, &v.Value, &v.Type, &v.Description, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}
	return &v, nil
}

func scanSiteSettingViewFromRows(rows pgx.Rows) (*domain.SiteSettingView, error) {
	var v domain.SiteSettingView
	err := rows.Scan(&v.ID, &v.Key, &v.Value, &v.Type, &v.Description, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
