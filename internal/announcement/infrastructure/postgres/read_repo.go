package postgres

import (
	"context"

	"gct/internal/announcement/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var readColumns = []string{
	"id", "title_uz", "title_ru", "title_en",
	"content_uz", "content_ru", "content_en",
	"published", "published_at", "priority",
	"start_date", "end_date", "created_at", "updated_at",
}

// AnnouncementReadRepo implements domain.AnnouncementReadRepository for the CQRS read side.
type AnnouncementReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewAnnouncementReadRepo creates a new AnnouncementReadRepo.
func NewAnnouncementReadRepo(pool *pgxpool.Pool) *AnnouncementReadRepo {
	return &AnnouncementReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// FindByID returns a single AnnouncementView by ID.
func (r *AnnouncementReadRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.AnnouncementView, error) {
	sql, args, err := r.builder.
		Select(readColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanAnnouncementView(row)
}

// List returns a paginated list of AnnouncementView with optional filters.
func (r *AnnouncementReadRepo) List(ctx context.Context, filter domain.AnnouncementFilter) ([]*domain.AnnouncementView, int64, error) {
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

	var views []*domain.AnnouncementView
	for rows.Next() {
		v, err := scanAnnouncementViewFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		views = append(views, v)
	}

	return views, total, nil
}

func scanAnnouncementView(row pgx.Row) (*domain.AnnouncementView, error) {
	var v domain.AnnouncementView
	err := row.Scan(&v.ID, &v.TitleUz, &v.TitleRu, &v.TitleEn,
		&v.ContentUz, &v.ContentRu, &v.ContentEn,
		&v.Published, &v.PublishedAt, &v.Priority,
		&v.StartDate, &v.EndDate, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}
	return &v, nil
}

func scanAnnouncementViewFromRows(rows pgx.Rows) (*domain.AnnouncementView, error) {
	var v domain.AnnouncementView
	err := rows.Scan(&v.ID, &v.TitleUz, &v.TitleRu, &v.TitleEn,
		&v.ContentUz, &v.ContentRu, &v.ContentEn,
		&v.Published, &v.PublishedAt, &v.Priority,
		&v.StartDate, &v.EndDate, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
