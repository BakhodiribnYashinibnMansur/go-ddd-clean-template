package postgres

import (
	"context"
	"time"

	translationentity "gct/internal/context/content/generic/translation/domain/entity"
	translationrepo "gct/internal/context/content/generic/translation/domain/repository"
	"gct/internal/kernel/consts"
	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableTranslations

var writeColumns = []string{
	"id", "entity_type", "entity_id", "lang_code", "data", "created_at", "updated_at",
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
func (r *TranslationWriteRepo) Save(ctx context.Context, q shareddomain.Querier, t *translationentity.Translation) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "TranslationWriteRepo.Save")
	defer func() { end(err) }()

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

	if _, err = q.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// FindByID retrieves a Translation aggregate by its ID.
func (r *TranslationWriteRepo) FindByID(ctx context.Context, id translationentity.TranslationID) (result *translationentity.Translation, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "TranslationWriteRepo.FindByID")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select(writeColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id.UUID()}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanTranslation(row)
}

// Update updates an existing Translation aggregate in the database.
func (r *TranslationWriteRepo) Update(ctx context.Context, q shareddomain.Querier, t *translationentity.Translation) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "TranslationWriteRepo.Update")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Update(tableName).
		Set("entity_type", t.Key()).
		Set("lang_code", t.Language()).
		Set("data", t.Value()).
		Set("entity_id", t.Group()).
		Set("updated_at", t.UpdatedAt()).
		Where(squirrel.Eq{"id": t.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = q.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// Delete removes a Translation by its ID.
func (r *TranslationWriteRepo) Delete(ctx context.Context, q shareddomain.Querier, id translationentity.TranslationID) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "TranslationWriteRepo.Delete")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Delete(tableName).
		Where(squirrel.Eq{"id": id.UUID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildDelete)
	}

	if _, err = q.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// List retrieves a paginated list of Translation aggregates with optional filters.
func (r *TranslationWriteRepo) List(ctx context.Context, filter translationrepo.TranslationFilter) (results []*translationentity.Translation, total int64, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "TranslationWriteRepo.List")
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

func applyFilters(conds squirrel.And, filter translationrepo.TranslationFilter) squirrel.And {
	if filter.Key != nil {
		conds = append(conds, squirrel.Eq{"entity_type": *filter.Key})
	}
	if filter.Language != nil {
		conds = append(conds, squirrel.Eq{"lang_code": *filter.Language})
	}
	if filter.Group != nil {
		conds = append(conds, squirrel.Eq{"entity_id": *filter.Group})
	}
	return conds
}

func scanTranslation(row pgx.Row) (*translationentity.Translation, error) {
	var (
		id         uuid.UUID
		entityType string
		entityID   uuid.UUID
		langCode   string
		data       []byte
		createdAt  time.Time
		updatedAt  time.Time
	)

	err := row.Scan(&id, &entityType, &entityID, &langCode, &data, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return translationentity.ReconstructTranslation(id, createdAt, updatedAt, entityType, langCode, string(data), entityID.String()), nil
}

func scanTranslationFromRows(rows pgx.Rows) (*translationentity.Translation, error) {
	var (
		id         uuid.UUID
		entityType string
		entityID   uuid.UUID
		langCode   string
		data       []byte
		createdAt  time.Time
		updatedAt  time.Time
	)

	err := rows.Scan(&id, &entityType, &entityID, &langCode, &data, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return translationentity.ReconstructTranslation(id, createdAt, updatedAt, entityType, langCode, string(data), entityID.String()), nil
}
