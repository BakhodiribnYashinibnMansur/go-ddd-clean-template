package postgres

import (
	"context"
	"time"

	"gct/internal/platform/domain/consts"
	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/pgxutil"
	"gct/internal/context/iam/usersetting/domain"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var readColumns = []string{
	"id", "user_id", "key", "value", "created_at", "updated_at",
}

// UserSettingReadRepo implements domain.UserSettingReadRepository for the CQRS read side.
type UserSettingReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewUserSettingReadRepo creates a new UserSettingReadRepo.
func NewUserSettingReadRepo(pool *pgxpool.Pool) *UserSettingReadRepo {
	return &UserSettingReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// FindByID returns a single UserSettingView by its ID.
func (r *UserSettingReadRepo) FindByID(ctx context.Context, id uuid.UUID) (result *domain.UserSettingView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "UserSettingReadRepo.FindByID")
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
	return scanUserSettingView(row)
}

// List returns a paginated list of UserSettingView with optional filters.
func (r *UserSettingReadRepo) List(ctx context.Context, filter domain.UserSettingFilter) (items []*domain.UserSettingView, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "UserSettingReadRepo.List")
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

	var views []*domain.UserSettingView
	for rows.Next() {
		v, err := scanUserSettingViewFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		views = append(views, v)
	}

	return views, total, nil
}

func applyFilters(conds squirrel.And, filter domain.UserSettingFilter) squirrel.And {
	if filter.UserID != nil {
		conds = append(conds, squirrel.Eq{"user_id": *filter.UserID})
	}
	if filter.Key != nil {
		conds = append(conds, squirrel.Eq{"key": *filter.Key})
	}
	return conds
}

func scanUserSettingView(row pgx.Row) (*domain.UserSettingView, error) {
	var (
		id        uuid.UUID
		userID    uuid.UUID
		key       string
		value     string
		createdAt time.Time
		updatedAt time.Time
	)

	err := row.Scan(&id, &userID, &key, &value, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return &domain.UserSettingView{
		ID:        id,
		UserID:    userID,
		Key:       key,
		Value:     value,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func scanUserSettingViewFromRows(rows pgx.Rows) (*domain.UserSettingView, error) {
	var (
		id        uuid.UUID
		userID    uuid.UUID
		key       string
		value     string
		createdAt time.Time
		updatedAt time.Time
	)

	err := rows.Scan(&id, &userID, &key, &value, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	return &domain.UserSettingView{
		ID:        id,
		UserID:    userID,
		Key:       key,
		Value:     value,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
