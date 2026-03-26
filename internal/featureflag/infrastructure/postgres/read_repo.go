package postgres

import (
	"context"
	"time"

	"gct/internal/featureflag/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var readColumns = []string{
	"id", "key", "name", "type", "value", "description", "is_active", "created_at", "updated_at",
}

// FeatureFlagReadRepo implements domain.FeatureFlagReadRepository for the CQRS read side.
type FeatureFlagReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewFeatureFlagReadRepo creates a new FeatureFlagReadRepo.
func NewFeatureFlagReadRepo(pool *pgxpool.Pool) *FeatureFlagReadRepo {
	return &FeatureFlagReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// FindByID returns a FeatureFlagView for the given ID.
func (r *FeatureFlagReadRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.FeatureFlagView, error) {
	sql, args, err := r.builder.
		Select(readColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanFeatureFlagView(row)
}

// List returns a paginated list of FeatureFlagView with optional filters.
func (r *FeatureFlagReadRepo) List(ctx context.Context, filter domain.FeatureFlagFilter) ([]*domain.FeatureFlagView, int64, error) {
	conds := squirrel.And{}
	if filter.Search != nil {
		conds = append(conds, squirrel.ILike{"name": "%" + *filter.Search + "%"})
	}
	if filter.Enabled != nil {
		conds = append(conds, squirrel.Eq{"is_active": *filter.Enabled})
	}

	// Count total.
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

	var views []*domain.FeatureFlagView
	for rows.Next() {
		v, err := scanFeatureFlagViewFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		views = append(views, v)
	}

	return views, total, nil
}

func scanFeatureFlagView(row pgx.Row) (*domain.FeatureFlagView, error) {
	var (
		id          uuid.UUID
		key         string
		name        string
		ffType      string
		value       string
		description string
		isActive    bool
		createdAt   time.Time
		updatedAt   time.Time
	)

	err := row.Scan(&id, &key, &name, &ffType, &value, &description, &isActive, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, map[string]any{"id": id})
	}

	_ = key
	_ = ffType
	_ = value

	return &domain.FeatureFlagView{
		ID:                id,
		Name:              name,
		Description:       description,
		Enabled:           isActive,
		RolloutPercentage: 0,
		CreatedAt:         createdAt.Format(time.RFC3339),
		UpdatedAt:         updatedAt.Format(time.RFC3339),
	}, nil
}

func scanFeatureFlagViewFromRows(rows pgx.Rows) (*domain.FeatureFlagView, error) {
	var (
		id          uuid.UUID
		key         string
		name        string
		ffType      string
		value       string
		description string
		isActive    bool
		createdAt   time.Time
		updatedAt   time.Time
	)

	err := rows.Scan(&id, &key, &name, &ffType, &value, &description, &isActive, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	_ = key
	_ = ffType
	_ = value

	return &domain.FeatureFlagView{
		ID:                id,
		Name:              name,
		Description:       description,
		Enabled:           isActive,
		RolloutPercentage: 0,
		CreatedAt:         createdAt.Format(time.RFC3339),
		UpdatedAt:         updatedAt.Format(time.RFC3339),
	}, nil
}
