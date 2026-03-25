package postgres

import (
	"context"
	"time"

	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/translation/domain"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableTranslations

var writeColumns = []string{
	"id", "key", "language", "value", "group_name", "created_at", "updated_at",
}

// TranslationWriteRepo implements domain.TranslationRepository using PostgreSQL.
type TranslationWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewTranslationWriteRepo creates a new TranslationWriteRepo.
func NewTranslationWriteRepo(pool *pgxpool.Pool) *TranslationWriteRepo {
	return &TranslationWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new Translation aggregate into the database.
func (r *TranslationWriteRepo) Save(ctx context.Context, t *domain.Translation) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			t.ID(), t.Key(), t.Language(), t.Value(), t.Group(),
			t.CreatedAt(), t.UpdatedAt(),
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

// FindByID retrieves a Translation aggregate by its ID.
func (r *TranslationWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Translation, error) {
	sql, args, err := r.builder.
		Select(writeColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanTranslation(row)
}

// Update updates an existing Translation aggregate in the database.
func (r *TranslationWriteRepo) Update(ctx context.Context, t *domain.Translation) error {
	sql, args, err := r.builder.
		Update(tableName).
		Set("key", t.Key()).
		Set("language", t.Language()).
		Set("value", t.Value()).
		Set("group_name", t.Group()).
		Set("updated_at", t.UpdatedAt()).
		Where(squirrel.Eq{"id": t.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// Delete removes a Translation by its ID.
func (r *TranslationWriteRepo) Delete(ctx context.Context, id uuid.UUID) error {
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

// List retrieves a paginated list of Translation aggregates with optional filters.
func (r *TranslationWriteRepo) List(ctx context.Context, filter domain.TranslationFilter) ([]*domain.Translation, int64, error) {
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

	var results []*domain.Translation
	for rows.Next() {
		t, err := scanTranslationFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		results = append(results, t)
	}

	return results, total, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func applyFilters(conds squirrel.And, filter domain.TranslationFilter) squirrel.And {
	if filter.Key != nil {
		conds = append(conds, squirrel.Eq{"key": *filter.Key})
	}
	if filter.Language != nil {
		conds = append(conds, squirrel.Eq{"language": *filter.Language})
	}
	if filter.Group != nil {
		conds = append(conds, squirrel.Eq{"group_name": *filter.Group})
	}
	return conds
}

func scanTranslation(row pgx.Row) (*domain.Translation, error) {
	var (
		id        uuid.UUID
		key       string
		language  string
		value     string
		group     string
		createdAt time.Time
		updatedAt time.Time
	)

	err := row.Scan(&id, &key, &language, &value, &group, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return domain.ReconstructTranslation(id, createdAt, updatedAt, key, language, value, group), nil
}

func scanTranslationFromRows(rows pgx.Rows) (*domain.Translation, error) {
	var (
		id        uuid.UUID
		key       string
		language  string
		value     string
		group     string
		createdAt time.Time
		updatedAt time.Time
	)

	err := rows.Scan(&id, &key, &language, &value, &group, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	return domain.ReconstructTranslation(id, createdAt, updatedAt, key, language, value, group), nil
}
