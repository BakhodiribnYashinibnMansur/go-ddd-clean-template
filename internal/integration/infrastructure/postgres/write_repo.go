package postgres

import (
	"context"
	"encoding/json"
	"time"

	"gct/internal/integration/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableIntegrations

var writeColumns = []string{
	"id", "name", "type", "api_key", "webhook_url", "enabled", "config", "created_at", "updated_at",
}

// IntegrationWriteRepo implements domain.IntegrationRepository using PostgreSQL.
type IntegrationWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewIntegrationWriteRepo creates a new IntegrationWriteRepo.
func NewIntegrationWriteRepo(pool *pgxpool.Pool) *IntegrationWriteRepo {
	return &IntegrationWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new Integration aggregate into the database.
func (r *IntegrationWriteRepo) Save(ctx context.Context, i *domain.Integration) error {
	configJSON, err := json.Marshal(i.Config())
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToMarshalJSON)
	}

	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			i.ID(),
			i.Name(),
			i.Type(),
			i.APIKey(),
			i.WebhookURL(),
			i.Enabled(),
			configJSON,
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

	return nil
}

// FindByID retrieves an Integration aggregate by ID.
func (r *IntegrationWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Integration, error) {
	sql, args, err := r.builder.
		Select(writeColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanIntegration(row)
}

// Update updates an Integration aggregate in the database.
func (r *IntegrationWriteRepo) Update(ctx context.Context, i *domain.Integration) error {
	configJSON, err := json.Marshal(i.Config())
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToMarshalJSON)
	}

	sql, args, err := r.builder.
		Update(tableName).
		Set("name", i.Name()).
		Set("type", i.Type()).
		Set("api_key", i.APIKey()).
		Set("webhook_url", i.WebhookURL()).
		Set("enabled", i.Enabled()).
		Set("config", configJSON).
		Set("updated_at", i.UpdatedAt()).
		Where(squirrel.Eq{"id": i.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// Delete removes an Integration by ID.
func (r *IntegrationWriteRepo) Delete(ctx context.Context, id uuid.UUID) error {
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
		id         uuid.UUID
		name       string
		intType    string
		apiKey     string
		webhookURL string
		enabled    bool
		configJSON []byte
		createdAt  time.Time
		updatedAt  time.Time
	)

	err := row.Scan(&id, &name, &intType, &apiKey, &webhookURL, &enabled, &configJSON, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, map[string]any{"id": id})
	}

	var config map[string]any
	if len(configJSON) > 0 {
		_ = json.Unmarshal(configJSON, &config)
	}

	return domain.ReconstructIntegration(id, createdAt, updatedAt, nil, name, intType, apiKey, webhookURL, enabled, config), nil
}
