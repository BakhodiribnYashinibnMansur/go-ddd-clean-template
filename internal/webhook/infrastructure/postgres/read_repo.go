package postgres

import (
	"context"
	"encoding/json"
	"time"

	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/webhook/domain"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var readColumns = []string{
	"id", "name", "url", "secret", "events", "enabled", "created_at", "updated_at",
}

// WebhookReadRepo implements domain.WebhookReadRepository for the CQRS read side.
type WebhookReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewWebhookReadRepo creates a new WebhookReadRepo.
func NewWebhookReadRepo(pool *pgxpool.Pool) *WebhookReadRepo {
	return &WebhookReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// FindByID returns a WebhookView for the given ID.
func (r *WebhookReadRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.WebhookView, error) {
	sql, args, err := r.builder.
		Select(readColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanWebhookView(row)
}

// List returns a paginated list of WebhookView with optional filters.
func (r *WebhookReadRepo) List(ctx context.Context, filter domain.WebhookFilter) ([]*domain.WebhookView, int64, error) {
	conds := squirrel.And{}
	if filter.Search != nil {
		conds = append(conds, squirrel.ILike{"name": "%" + *filter.Search + "%"})
	}
	if filter.Enabled != nil {
		conds = append(conds, squirrel.Eq{"enabled": *filter.Enabled})
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

	var views []*domain.WebhookView
	for rows.Next() {
		v, err := scanWebhookViewFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		views = append(views, v)
	}

	return views, total, nil
}

func scanWebhookView(row pgx.Row) (*domain.WebhookView, error) {
	var (
		id         uuid.UUID
		name       string
		url        string
		secret     string
		eventsJSON []byte
		enabled    bool
		createdAt  time.Time
		updatedAt  time.Time
	)

	err := row.Scan(&id, &name, &url, &secret, &eventsJSON, &enabled, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, map[string]any{"id": id})
	}

	var events []string
	if len(eventsJSON) > 0 {
		_ = json.Unmarshal(eventsJSON, &events)
	}

	return &domain.WebhookView{
		ID:        id,
		Name:      name,
		URL:       url,
		Secret:    secret,
		Events:    events,
		Enabled:   enabled,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func scanWebhookViewFromRows(rows pgx.Rows) (*domain.WebhookView, error) {
	var (
		id         uuid.UUID
		name       string
		url        string
		secret     string
		eventsJSON []byte
		enabled    bool
		createdAt  time.Time
		updatedAt  time.Time
	)

	err := rows.Scan(&id, &name, &url, &secret, &eventsJSON, &enabled, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	var events []string
	if len(eventsJSON) > 0 {
		_ = json.Unmarshal(eventsJSON, &events)
	}

	return &domain.WebhookView{
		ID:        id,
		Name:      name,
		URL:       url,
		Secret:    secret,
		Events:    events,
		Enabled:   enabled,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
