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

const tableName = consts.TableWebhooks

var writeColumns = []string{
	"id", "name", "url", "secret", "events", "enabled", "created_at", "updated_at",
}

// WebhookWriteRepo implements domain.WebhookRepository using PostgreSQL.
type WebhookWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewWebhookWriteRepo creates a new WebhookWriteRepo.
func NewWebhookWriteRepo(pool *pgxpool.Pool) *WebhookWriteRepo {
	return &WebhookWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new Webhook aggregate into the database.
func (r *WebhookWriteRepo) Save(ctx context.Context, w *domain.Webhook) error {
	eventsJSON, err := json.Marshal(w.Events_())
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToMarshalJSON)
	}

	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			w.ID(),
			w.Name(),
			w.URL(),
			w.Secret(),
			eventsJSON,
			w.Enabled(),
			w.CreatedAt(),
			w.UpdatedAt(),
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

// FindByID retrieves a Webhook aggregate by ID.
func (r *WebhookWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Webhook, error) {
	sql, args, err := r.builder.
		Select(writeColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanWebhook(row)
}

// Update updates a Webhook aggregate in the database.
func (r *WebhookWriteRepo) Update(ctx context.Context, w *domain.Webhook) error {
	eventsJSON, err := json.Marshal(w.Events_())
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToMarshalJSON)
	}

	sql, args, err := r.builder.
		Update(tableName).
		Set("name", w.Name()).
		Set("url", w.URL()).
		Set("secret", w.Secret()).
		Set("events", eventsJSON).
		Set("enabled", w.Enabled()).
		Set("updated_at", w.UpdatedAt()).
		Where(squirrel.Eq{"id": w.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// Delete removes a Webhook by ID.
func (r *WebhookWriteRepo) Delete(ctx context.Context, id uuid.UUID) error {
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

func scanWebhook(row pgx.Row) (*domain.Webhook, error) {
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

	return domain.ReconstructWebhook(id, createdAt, updatedAt, nil, name, url, secret, events, enabled), nil
}
