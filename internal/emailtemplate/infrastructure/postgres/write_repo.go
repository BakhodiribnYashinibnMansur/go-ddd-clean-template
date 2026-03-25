package postgres

import (
	"context"
	"encoding/json"
	"time"

	"gct/internal/emailtemplate/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableEmailTemplates

var writeColumns = []string{
	"id", "name", "subject", "html_body", "text_body", "variables", "created_at", "updated_at",
}

// EmailTemplateWriteRepo implements domain.EmailTemplateRepository using PostgreSQL.
type EmailTemplateWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewEmailTemplateWriteRepo creates a new EmailTemplateWriteRepo.
func NewEmailTemplateWriteRepo(pool *pgxpool.Pool) *EmailTemplateWriteRepo {
	return &EmailTemplateWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new EmailTemplate aggregate into the database.
func (r *EmailTemplateWriteRepo) Save(ctx context.Context, et *domain.EmailTemplate) error {
	varsJSON, err := json.Marshal(et.Variables())
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToMarshalJSON)
	}

	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			et.ID(),
			et.Name(),
			et.Subject(),
			et.HTMLBody(),
			et.TextBody(),
			varsJSON,
			et.CreatedAt(),
			et.UpdatedAt(),
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

// FindByID retrieves an EmailTemplate aggregate by ID.
func (r *EmailTemplateWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.EmailTemplate, error) {
	sql, args, err := r.builder.
		Select(writeColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanEmailTemplate(row)
}

// Update updates an EmailTemplate aggregate in the database.
func (r *EmailTemplateWriteRepo) Update(ctx context.Context, et *domain.EmailTemplate) error {
	varsJSON, err := json.Marshal(et.Variables())
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToMarshalJSON)
	}

	sql, args, err := r.builder.
		Update(tableName).
		Set("name", et.Name()).
		Set("subject", et.Subject()).
		Set("html_body", et.HTMLBody()).
		Set("text_body", et.TextBody()).
		Set("variables", varsJSON).
		Set("updated_at", et.UpdatedAt()).
		Where(squirrel.Eq{"id": et.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// Delete removes an EmailTemplate by ID.
func (r *EmailTemplateWriteRepo) Delete(ctx context.Context, id uuid.UUID) error {
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

func scanEmailTemplate(row pgx.Row) (*domain.EmailTemplate, error) {
	var (
		id        uuid.UUID
		name      string
		subject   string
		htmlBody  string
		textBody  string
		varsJSON  []byte
		createdAt time.Time
		updatedAt time.Time
	)

	err := row.Scan(&id, &name, &subject, &htmlBody, &textBody, &varsJSON, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, map[string]any{"id": id})
	}

	var variables []string
	if len(varsJSON) > 0 {
		_ = json.Unmarshal(varsJSON, &variables)
	}

	return domain.ReconstructEmailTemplate(id, createdAt, updatedAt, nil, name, subject, htmlBody, textBody, variables), nil
}
