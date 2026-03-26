package postgres

import (
	"context"
	"time"

	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/sitesetting/domain"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableSiteSetting

var writeColumns = []string{
	"id", "key", "value", "value_type", "description", "created_at", "updated_at",
}

// SiteSettingWriteRepo implements domain.SiteSettingRepository using PostgreSQL.
type SiteSettingWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewSiteSettingWriteRepo creates a new SiteSettingWriteRepo.
func NewSiteSettingWriteRepo(pool *pgxpool.Pool) *SiteSettingWriteRepo {
	return &SiteSettingWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new SiteSetting aggregate into the database.
func (r *SiteSettingWriteRepo) Save(ctx context.Context, s *domain.SiteSetting) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			s.ID(), s.Key(), s.Value(), s.Type(), s.Description(),
			s.CreatedAt(), s.UpdatedAt(),
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

// FindByID retrieves a SiteSetting aggregate by its ID.
func (r *SiteSettingWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.SiteSetting, error) {
	sql, args, err := r.builder.
		Select(writeColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanSiteSetting(row)
}

// Update updates an existing SiteSetting aggregate in the database.
func (r *SiteSettingWriteRepo) Update(ctx context.Context, s *domain.SiteSetting) error {
	sql, args, err := r.builder.
		Update(tableName).
		Set("key", s.Key()).
		Set("value", s.Value()).
		Set("value_type", s.Type()).
		Set("description", s.Description()).
		Set("updated_at", s.UpdatedAt()).
		Where(squirrel.Eq{"id": s.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// Delete removes a SiteSetting by its ID.
func (r *SiteSettingWriteRepo) Delete(ctx context.Context, id uuid.UUID) error {
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

// List retrieves a paginated list of SiteSetting aggregates with optional filters.
func (r *SiteSettingWriteRepo) List(ctx context.Context, filter domain.SiteSettingFilter) ([]*domain.SiteSetting, int64, error) {
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

	var results []*domain.SiteSetting
	for rows.Next() {
		s, err := scanSiteSettingFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		results = append(results, s)
	}

	return results, total, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func applyFilters(conds squirrel.And, filter domain.SiteSettingFilter) squirrel.And {
	if filter.Key != nil {
		conds = append(conds, squirrel.Eq{"key": *filter.Key})
	}
	if filter.Type != nil {
		conds = append(conds, squirrel.Eq{"value_type": *filter.Type})
	}
	return conds
}

func scanSiteSetting(row pgx.Row) (*domain.SiteSetting, error) {
	var (
		id          uuid.UUID
		key         string
		value       string
		sType       string
		description string
		createdAt   time.Time
		updatedAt   time.Time
	)

	err := row.Scan(&id, &key, &value, &sType, &description, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return domain.ReconstructSiteSetting(id, createdAt, updatedAt, key, value, sType, description), nil
}

func scanSiteSettingFromRows(rows pgx.Rows) (*domain.SiteSetting, error) {
	var (
		id          uuid.UUID
		key         string
		value       string
		sType       string
		description string
		createdAt   time.Time
		updatedAt   time.Time
	)

	err := rows.Scan(&id, &key, &value, &sType, &description, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	return domain.ReconstructSiteSetting(id, createdAt, updatedAt, key, value, sType, description), nil
}
