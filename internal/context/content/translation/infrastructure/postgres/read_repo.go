package postgres

import (
	"context"

	"gct/internal/kernel/consts"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/pgxutil"
	"gct/internal/context/content/translation/domain"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var readColumns = []string{
	"id", "entity_type", "entity_id::text", "lang_code", "data::text", "created_at", "updated_at",
}

// TranslationReadRepo implements domain.TranslationReadRepository for the CQRS read side.
type TranslationReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewTranslationReadRepo creates a new TranslationReadRepo.
func NewTranslationReadRepo(pool *pgxpool.Pool) *TranslationReadRepo {
	return &TranslationReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// FindByID returns a single TranslationView by ID.
func (r *TranslationReadRepo) FindByID(ctx context.Context, id uuid.UUID) (result *domain.TranslationView, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "TranslationReadRepo.FindByID")
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
	return scanTranslationView(row)
}

// List returns a paginated list of TranslationView with optional filters.
func (r *TranslationReadRepo) List(ctx context.Context, filter domain.TranslationFilter) (views []*domain.TranslationView, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "TranslationReadRepo.List")
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
		v, err := scanTranslationViewFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		views = append(views, v)
	}

	return views, total, nil
}

func scanTranslationView(row pgx.Row) (*domain.TranslationView, error) {
	var v domain.TranslationView
	err := row.Scan(&v.ID, &v.Key, &v.Language, &v.Value, &v.Group, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}
	return &v, nil
}

func scanTranslationViewFromRows(rows pgx.Rows) (*domain.TranslationView, error) {
	var v domain.TranslationView
	err := rows.Scan(&v.ID, &v.Key, &v.Language, &v.Value, &v.Group, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}
	return &v, nil
}
