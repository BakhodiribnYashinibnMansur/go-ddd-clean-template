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

const tableName = consts.TableFeatureFlags

var writeColumns = []string{
	"id", "key", "name", "type", "value", "description", "is_active", "created_at", "updated_at",
}

// FeatureFlagWriteRepo implements domain.FeatureFlagRepository using PostgreSQL.
type FeatureFlagWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewFeatureFlagWriteRepo creates a new FeatureFlagWriteRepo.
func NewFeatureFlagWriteRepo(pool *pgxpool.Pool) *FeatureFlagWriteRepo {
	return &FeatureFlagWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new FeatureFlag aggregate into the database.
func (r *FeatureFlagWriteRepo) Save(ctx context.Context, ff *domain.FeatureFlag) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			ff.ID(),
			ff.Name(),
			ff.Name(),
			"bool",
			"",
			ff.Description(),
			ff.Enabled(),
			ff.CreatedAt(),
			ff.UpdatedAt(),
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

// FindByID retrieves a FeatureFlag aggregate by ID.
func (r *FeatureFlagWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.FeatureFlag, error) {
	sql, args, err := r.builder.
		Select(writeColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanFeatureFlag(row)
}

// Update updates a FeatureFlag aggregate in the database.
func (r *FeatureFlagWriteRepo) Update(ctx context.Context, ff *domain.FeatureFlag) error {
	sql, args, err := r.builder.
		Update(tableName).
		Set("name", ff.Name()).
		Set("description", ff.Description()).
		Set("is_active", ff.Enabled()).
		Set("updated_at", ff.UpdatedAt()).
		Where(squirrel.Eq{"id": ff.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// Delete removes a FeatureFlag by ID.
func (r *FeatureFlagWriteRepo) Delete(ctx context.Context, id uuid.UUID) error {
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

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func scanFeatureFlag(row pgx.Row) (*domain.FeatureFlag, error) {
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

	return domain.ReconstructFeatureFlag(id, createdAt, updatedAt, nil, name, description, isActive, 0), nil
}
