package postgres

import (
	"context"
	"time"

	"gct/internal/context/content/announcement/domain"
	shared "gct/internal/platform/domain"
	"gct/internal/platform/domain/consts"
	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableAnnouncements

var writeColumns = []string{
	"id", "title", "content", "type", "is_active",
	"priority", "starts_at", "ends_at", "created_at", "updated_at",
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
func (r *AnnouncementWriteRepo) Save(ctx context.Context, a *domain.Announcement) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "AnnouncementWriteRepo.Save")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			a.ID(),
			a.Title().Uz,
			a.Content().Uz,
			"info",
			a.Published(),
			a.Priority(),
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
func (r *AnnouncementWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (result *domain.Announcement, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "AnnouncementWriteRepo.FindByID")
	defer func() { end(err) }()

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
func (r *AnnouncementWriteRepo) Update(ctx context.Context, a *domain.Announcement) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "AnnouncementWriteRepo.Update")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Update(tableName).
		Set("title", a.Title().Uz).
		Set("content", a.Content().Uz).
		Set("is_active", a.Published()).
		Set("priority", a.Priority()).
		Set("starts_at", a.StartDate()).
		Set("ends_at", a.EndDate()).
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
func (r *AnnouncementWriteRepo) Delete(ctx context.Context, id uuid.UUID) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "AnnouncementWriteRepo.Delete")
	defer func() { end(err) }()

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
func (r *AnnouncementWriteRepo) List(ctx context.Context, filter domain.AnnouncementFilter) (results []*domain.Announcement, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "AnnouncementWriteRepo.List")
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
		conds = append(conds, squirrel.Eq{"is_active": *filter.Published})
	}
	return conds
}

func scanAnnouncement(row pgx.Row) (*domain.Announcement, error) {
	var (
		id        uuid.UUID
		title     string
		content   string
		aType     string
		isActive  bool
		priority  int
		startsAt  *time.Time
		endsAt    *time.Time
		createdAt time.Time
		updatedAt time.Time
	)

	err := row.Scan(&id, &title, &content, &aType, &isActive, &priority, &startsAt, &endsAt, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	_ = aType

	return domain.ReconstructAnnouncement(
		id, createdAt, updatedAt,
		shared.Lang{Uz: title, Ru: title, En: title},
		shared.Lang{Uz: content, Ru: content, En: content},
		isActive, nil, priority, startsAt, endsAt,
	), nil
}

func scanAnnouncementFromRows(rows pgx.Rows) (*domain.Announcement, error) {
	var (
		id        uuid.UUID
		title     string
		content   string
		aType     string
		isActive  bool
		priority  int
		startsAt  *time.Time
		endsAt    *time.Time
		createdAt time.Time
		updatedAt time.Time
	)

	err := rows.Scan(&id, &title, &content, &aType, &isActive, &priority, &startsAt, &endsAt, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	_ = aType

	return domain.ReconstructAnnouncement(
		id, createdAt, updatedAt,
		shared.Lang{Uz: title, Ru: title, En: title},
		shared.Lang{Uz: content, Ru: content, En: content},
		isActive, nil, priority, startsAt, endsAt,
	), nil
}
