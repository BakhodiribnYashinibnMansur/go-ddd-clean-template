package postgres

import (
	"context"
	"time"

	"gct/internal/integration/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/metadata"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableIntegrations

var writeColumns = []string{
	"id", "name", "description", "base_url", "is_active", "created_at", "updated_at",
}

// IntegrationWriteRepo implements domain.IntegrationRepository using PostgreSQL.
type IntegrationWriteRepo struct {
	pool     *pgxpool.Pool
	builder  squirrel.StatementBuilderType
	metadata *metadata.GenericMetadataRepo
}

// NewIntegrationWriteRepo creates a new IntegrationWriteRepo.
func NewIntegrationWriteRepo(pool *pgxpool.Pool) *IntegrationWriteRepo {
	return &IntegrationWriteRepo{
		pool:     pool,
		builder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		metadata: metadata.NewGenericMetadataRepo(pool),
	}
}

// Save inserts a new Integration aggregate into the database.
func (r *IntegrationWriteRepo) Save(ctx context.Context, i *domain.Integration) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "IntegrationWriteRepo.Save")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			i.ID(),
			i.Name(),
			i.Type(),
			i.WebhookURL(),
			i.Enabled(),
			i.CreatedAt(),
			i.UpdatedAt(),
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	if err := r.metadata.SetMany(ctx, metadata.EntityTypeIntegrationConfig, i.ID(), i.Config()); err != nil {
		return err
	}

	return nil
}

// FindByID retrieves an Integration aggregate by ID.
func (r *IntegrationWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (result *domain.Integration, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "IntegrationWriteRepo.FindByID")
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
	entity, err := scanIntegration(row)
	if err != nil {
		return nil, err
	}

	config, err := r.metadata.GetAll(ctx, metadata.EntityTypeIntegrationConfig, entity.ID())
	if err != nil {
		return nil, err
	}

	return domain.ReconstructIntegration(
		entity.ID(), entity.CreatedAt(), entity.UpdatedAt(), nil,
		entity.Name(), entity.Type(), entity.APIKey(), entity.WebhookURL(),
		entity.Enabled(), config,
	), nil
}

// Update updates an Integration aggregate in the database.
func (r *IntegrationWriteRepo) Update(ctx context.Context, i *domain.Integration) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "IntegrationWriteRepo.Update")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Update(tableName).
		Set("name", i.Name()).
		Set("description", i.Type()).
		Set("base_url", i.WebhookURL()).
		Set("is_active", i.Enabled()).
		Set("updated_at", i.UpdatedAt()).
		Where(squirrel.Eq{"id": i.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	if err := r.metadata.SetMany(ctx, metadata.EntityTypeIntegrationConfig, i.ID(), i.Config()); err != nil {
		return err
	}

	return nil
}

// Delete removes an Integration by ID.
func (r *IntegrationWriteRepo) Delete(ctx context.Context, id uuid.UUID) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "IntegrationWriteRepo.Delete")
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

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func scanIntegration(row pgx.Row) (*domain.Integration, error) {
	var (
		id        uuid.UUID
		name      string
		description *string
		baseURL   string
		isActive  bool
		createdAt time.Time
		updatedAt time.Time
	)

	err := row.Scan(&id, &name, &description, &baseURL, &isActive, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, map[string]any{"id": id})
	}

	desc := ""
	if description != nil {
		desc = *description
	}

	return domain.ReconstructIntegration(id, createdAt, updatedAt, nil, name, desc, "", baseURL, isActive, nil), nil
}
