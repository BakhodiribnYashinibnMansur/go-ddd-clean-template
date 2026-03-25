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

var readColumns = []string{
	"id", "name", "subject", "html_body", "text_body", "variables", "created_at", "updated_at",
}

// EmailTemplateReadRepo implements domain.EmailTemplateReadRepository for the CQRS read side.
type EmailTemplateReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewEmailTemplateReadRepo creates a new EmailTemplateReadRepo.
func NewEmailTemplateReadRepo(pool *pgxpool.Pool) *EmailTemplateReadRepo {
	return &EmailTemplateReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// FindByID returns an EmailTemplateView for the given ID.
func (r *EmailTemplateReadRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.EmailTemplateView, error) {
	sql, args, err := r.builder.
		Select(readColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanEmailTemplateView(row)
}

// List returns a paginated list of EmailTemplateView with optional filters.
func (r *EmailTemplateReadRepo) List(ctx context.Context, filter domain.EmailTemplateFilter) ([]*domain.EmailTemplateView, int64, error) {
	conds := squirrel.And{}
	if filter.Search != nil {
		conds = append(conds, squirrel.ILike{"name": "%" + *filter.Search + "%"})
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

	var views []*domain.EmailTemplateView
	for rows.Next() {
		v, err := scanEmailTemplateViewFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		views = append(views, v)
	}

	return views, total, nil
}

func scanEmailTemplateView(row pgx.Row) (*domain.EmailTemplateView, error) {
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

	return &domain.EmailTemplateView{
		ID:        id,
		Name:      name,
		Subject:   subject,
		HTMLBody:  htmlBody,
		TextBody:  textBody,
		Variables: variables,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func scanEmailTemplateViewFromRows(rows pgx.Rows) (*domain.EmailTemplateView, error) {
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

	err := rows.Scan(&id, &name, &subject, &htmlBody, &textBody, &varsJSON, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	var variables []string
	if len(varsJSON) > 0 {
		_ = json.Unmarshal(varsJSON, &variables)
	}

	return &domain.EmailTemplateView{
		ID:        id,
		Name:      name,
		Subject:   subject,
		HTMLBody:  htmlBody,
		TextBody:  textBody,
		Variables: variables,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
