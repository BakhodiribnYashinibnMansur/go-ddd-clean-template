package postgres

import (
	"context"

	siteentity "gct/internal/context/admin/supporting/sitesetting/domain/entity"
	siterepo "gct/internal/context/admin/supporting/sitesetting/domain/repository"
	"gct/internal/kernel/consts"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var readColumns = []string{
	"id", "key", "value", "value_type", "description", "created_at", "updated_at",
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
func (r *SiteSettingReadRepo) FindByID(ctx context.Context, id siteentity.SiteSettingID) (result *siterepo.SiteSettingView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "SiteSettingReadRepo.FindByID")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select(readColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id.UUID()}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanSiteSettingView(row)
}

// List returns a paginated list of SiteSettingView with optional filters.
func (r *SiteSettingReadRepo) List(ctx context.Context, filter siterepo.SiteSettingFilter) (views []*siterepo.SiteSettingView, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "SiteSettingReadRepo.List")
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
		v, err := scanSiteSettingViewFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		views = append(views, v)
	}

	return views, total, nil
}

func scanSiteSettingView(row pgx.Row) (*siterepo.SiteSettingView, error) {
	var (
		v     siterepo.SiteSettingView
		rawID uuid.UUID
	)
	err := row.Scan(&rawID, &v.Key, &v.Value, &v.Type, &v.Description, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}
	v.ID = siteentity.SiteSettingID(rawID)
	return &v, nil
}

func scanSiteSettingViewFromRows(rows pgx.Rows) (*siterepo.SiteSettingView, error) {
	var (
		v     siterepo.SiteSettingView
		rawID uuid.UUID
	)
	err := rows.Scan(&rawID, &v.Key, &v.Value, &v.Type, &v.Description, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}
	v.ID = siteentity.SiteSettingID(rawID)
	return &v, nil
}
