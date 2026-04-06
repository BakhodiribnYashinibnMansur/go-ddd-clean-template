package postgres

import (
	"context"
	"time"

	announceentity "gct/internal/context/content/supporting/announcement/domain/entity"
	announcerepo "gct/internal/context/content/supporting/announcement/domain/repository"
	"gct/internal/kernel/consts"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var readColumns = []string{
	"id", "title", "content", "type", "is_active",
	"priority", "starts_at", "ends_at", "created_at", "updated_at",
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
func (r *AnnouncementReadRepo) FindByID(ctx context.Context, id announceentity.AnnouncementID) (result *announcerepo.AnnouncementView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "AnnouncementReadRepo.FindByID")
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
	return scanAnnouncementView(row)
}

// List returns a paginated list of AnnouncementView with optional filters.
func (r *AnnouncementReadRepo) List(ctx context.Context, filter announcerepo.AnnouncementFilter) (views []*announcerepo.AnnouncementView, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "AnnouncementReadRepo.List")
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
		v, err := scanAnnouncementViewFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		views = append(views, v)
	}

	return views, total, nil
}

func scanAnnouncementView(row pgx.Row) (*announcerepo.AnnouncementView, error) {
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
	return &announcerepo.AnnouncementView{
		ID:      announceentity.AnnouncementID(id),
		TitleUz: title, TitleRu: title, TitleEn: title,
		ContentUz: content, ContentRu: content, ContentEn: content,
		Published: isActive,
		Priority:  priority,
		StartDate: startsAt,
		EndDate:   endsAt,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func scanAnnouncementViewFromRows(rows pgx.Rows) (*announcerepo.AnnouncementView, error) {
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
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}
	_ = aType
	return &announcerepo.AnnouncementView{
		ID:      announceentity.AnnouncementID(id),
		TitleUz: title, TitleRu: title, TitleEn: title,
		ContentUz: content, ContentRu: content, ContentEn: content,
		Published: isActive,
		Priority:  priority,
		StartDate: startsAt,
		EndDate:   endsAt,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
