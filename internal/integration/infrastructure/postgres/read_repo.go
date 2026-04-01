package postgres

import (
	"context"
	"time"

	"gct/internal/integration/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/metadata"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableAPIKeys = consts.TableAPIKeys

var readColumns = []string{
	"id", "name", "description", "base_url", "is_active", "created_at", "updated_at",
}

// IntegrationReadRepo implements domain.IntegrationReadRepository for the CQRS read side.
type IntegrationReadRepo struct {
	pool     *pgxpool.Pool
	builder  squirrel.StatementBuilderType
	metadata *metadata.GenericMetadataRepo
}

// NewIntegrationReadRepo creates a new IntegrationReadRepo.
func NewIntegrationReadRepo(pool *pgxpool.Pool) *IntegrationReadRepo {
	return &IntegrationReadRepo{
		pool:     pool,
		builder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		metadata: metadata.NewGenericMetadataRepo(pool),
	}
}

// FindByID returns an IntegrationView for the given ID.
func (r *IntegrationReadRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.IntegrationView, error) {
	sql, args, err := r.builder.
		Select(readColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	view, err := scanIntegrationView(row)
	if err != nil {
		return nil, err
	}

	config, err := r.metadata.GetAll(ctx, metadata.EntityTypeIntegrationConfig, view.ID)
	if err != nil {
		return nil, err
	}
	view.Config = config

	return view, nil
}

// List returns a paginated list of IntegrationView with optional filters.
func (r *IntegrationReadRepo) List(ctx context.Context, filter domain.IntegrationFilter) ([]*domain.IntegrationView, int64, error) {
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

	var views []*domain.IntegrationView
	for rows.Next() {
		v, err := scanIntegrationViewFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		views = append(views, v)
	}

	// Load metadata for each integration.
	for _, v := range views {
		config, err := r.metadata.GetAll(ctx, metadata.EntityTypeIntegrationConfig, v.ID)
		if err != nil {
			return nil, 0, err
		}
		v.Config = config
	}

	return views, total, nil
}

func scanIntegrationView(row pgx.Row) (*domain.IntegrationView, error) {
	var (
		id          uuid.UUID
		name        string
		description *string
		baseURL     string
		isActive    bool
		createdAt   time.Time
		updatedAt   time.Time
	)

	err := row.Scan(&id, &name, &description, &baseURL, &isActive, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, map[string]any{"id": id})
	}

	desc := ""
	if description != nil {
		desc = *description
	}

	return &domain.IntegrationView{
		ID:         id,
		Name:       name,
		Type:       desc,
		APIKey:     "",
		WebhookURL: baseURL,
		Enabled:    isActive,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

func scanIntegrationViewFromRows(rows pgx.Rows) (*domain.IntegrationView, error) {
	var (
		id          uuid.UUID
		name        string
		description *string
		baseURL     string
		isActive    bool
		createdAt   time.Time
		updatedAt   time.Time
	)

	err := rows.Scan(&id, &name, &description, &baseURL, &isActive, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	desc := ""
	if description != nil {
		desc = *description
	}

	return &domain.IntegrationView{
		ID:         id,
		Name:       name,
		Type:       desc,
		APIKey:     "",
		WebhookURL: baseURL,
		Enabled:    isActive,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

// FindByAPIKey returns an IntegrationAPIKeyView for the given API key string.
func (r *IntegrationReadRepo) FindByAPIKey(ctx context.Context, apiKey string) (*domain.IntegrationAPIKeyView, error) {
	sql, args, err := r.builder.
		Select("id", "integration_id", "key", "is_active").
		From(tableAPIKeys).
		Where(squirrel.Eq{"key": apiKey}).
		Where(squirrel.Eq{"is_active": true}).
		Where(squirrel.Eq{"deleted_at": nil}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	var view domain.IntegrationAPIKeyView
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&view.ID,
		&view.IntegrationID,
		&view.Key,
		&view.Active,
	)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableAPIKeys, nil)
	}

	return &view, nil
}
