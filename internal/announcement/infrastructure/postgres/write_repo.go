package postgres

import (
	"context"
	"time"

	"gct/internal/announcement/domain"
	shared "gct/internal/shared/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableAnnouncements

var writeColumns = []string{
	"id", "title_uz", "title_ru", "title_en",
	"content_uz", "content_ru", "content_en",
	"published", "published_at", "priority",
	"start_date", "end_date", "created_at", "updated_at",
}

// AnnouncementWriteRepo implements domain.AnnouncementRepository using PostgreSQL.
type AnnouncementWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewAnnouncementWriteRepo creates a new AnnouncementWriteRepo.
func NewAnnouncementWriteRepo(pool *pgxpool.Pool) *AnnouncementWriteRepo {
	return &AnnouncementWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new Announcement aggregate into the database.
func (r *AnnouncementWriteRepo) Save(ctx context.Context, a *domain.Announcement) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			a.ID(),
			a.Title().Uz, a.Title().Ru, a.Title().En,
			a.Content().Uz, a.Content().Ru, a.Content().En,
			a.Published(), a.PublishedAt(), a.Priority(),
			a.StartDate(), a.EndDate(),
			a.CreatedAt(), a.UpdatedAt(),
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

// FindByID retrieves an Announcement aggregate by its ID.
func (r *AnnouncementWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Announcement, error) {
	sql, args, err := r.builder.
		Select(writeColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanAnnouncement(row)
}

// Update updates an existing Announcement aggregate in the database.
func (r *AnnouncementWriteRepo) Update(ctx context.Context, a *domain.Announcement) error {
	sql, args, err := r.builder.
		Update(tableName).
		Set("title_uz", a.Title().Uz).
		Set("title_ru", a.Title().Ru).
		Set("title_en", a.Title().En).
		Set("content_uz", a.Content().Uz).
		Set("content_ru", a.Content().Ru).
		Set("content_en", a.Content().En).
		Set("published", a.Published()).
		Set("published_at", a.PublishedAt()).
		Set("priority", a.Priority()).
		Set("start_date", a.StartDate()).
		Set("end_date", a.EndDate()).
		Set("updated_at", a.UpdatedAt()).
		Where(squirrel.Eq{"id": a.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// Delete removes an Announcement by its ID.
func (r *AnnouncementWriteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	sql, args, err := r.builder.
		Delete(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// List retrieves a paginated list of Announcement aggregates with optional filters.
func (r *AnnouncementWriteRepo) List(ctx context.Context, filter domain.AnnouncementFilter) ([]*domain.Announcement, int64, error) {
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

	var results []*domain.Announcement
	for rows.Next() {
		a, err := scanAnnouncementFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		results = append(results, a)
	}

	return results, total, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func applyFilters(conds squirrel.And, filter domain.AnnouncementFilter) squirrel.And {
	if filter.Published != nil {
		conds = append(conds, squirrel.Eq{"published": *filter.Published})
	}
	if filter.Priority != nil {
		conds = append(conds, squirrel.Eq{"priority": *filter.Priority})
	}
	return conds
}

func scanAnnouncement(row pgx.Row) (*domain.Announcement, error) {
	var (
		id          uuid.UUID
		titleUz     string
		titleRu     string
		titleEn     string
		contentUz   string
		contentRu   string
		contentEn   string
		published   bool
		publishedAt *time.Time
		priority    int
		startDate   *time.Time
		endDate     *time.Time
		createdAt   time.Time
		updatedAt   time.Time
	)

	err := row.Scan(&id, &titleUz, &titleRu, &titleEn,
		&contentUz, &contentRu, &contentEn,
		&published, &publishedAt, &priority,
		&startDate, &endDate, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return domain.ReconstructAnnouncement(
		id, createdAt, updatedAt,
		shared.Lang{Uz: titleUz, Ru: titleRu, En: titleEn},
		shared.Lang{Uz: contentUz, Ru: contentRu, En: contentEn},
		published, publishedAt, priority, startDate, endDate,
	), nil
}

func scanAnnouncementFromRows(rows pgx.Rows) (*domain.Announcement, error) {
	var (
		id          uuid.UUID
		titleUz     string
		titleRu     string
		titleEn     string
		contentUz   string
		contentRu   string
		contentEn   string
		published   bool
		publishedAt *time.Time
		priority    int
		startDate   *time.Time
		endDate     *time.Time
		createdAt   time.Time
		updatedAt   time.Time
	)

	err := rows.Scan(&id, &titleUz, &titleRu, &titleEn,
		&contentUz, &contentRu, &contentEn,
		&published, &publishedAt, &priority,
		&startDate, &endDate, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	return domain.ReconstructAnnouncement(
		id, createdAt, updatedAt,
		shared.Lang{Uz: titleUz, Ru: titleRu, En: titleEn},
		shared.Lang{Uz: contentUz, Ru: contentRu, En: contentEn},
		published, publishedAt, priority, startDate, endDate,
	), nil
}
