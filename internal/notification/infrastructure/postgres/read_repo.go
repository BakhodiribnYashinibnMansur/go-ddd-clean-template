package postgres

import (
	"context"
	"time"

	"gct/internal/notification/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var readColumns = []string{
	"id", "title", "body", "type", "target_type", "is_active", "created_at", "updated_at",
}

// NotificationReadRepo implements domain.NotificationReadRepository for the CQRS read side.
type NotificationReadRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewNotificationReadRepo creates a new NotificationReadRepo.
func NewNotificationReadRepo(pool *pgxpool.Pool) *NotificationReadRepo {
	return &NotificationReadRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// FindByID returns a NotificationView for the given ID.
func (r *NotificationReadRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.NotificationView, error) {
	sql, args, err := r.builder.
		Select(readColumns...).
		From(tableName).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanNotificationView(row)
}

// List returns a paginated list of NotificationView with optional filters.
func (r *NotificationReadRepo) List(ctx context.Context, filter domain.NotificationFilter) ([]*domain.NotificationView, int64, error) {
	conds := squirrel.And{}
	if filter.Type != nil {
		conds = append(conds, squirrel.Eq{"type": *filter.Type})
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

	var views []*domain.NotificationView
	for rows.Next() {
		v, err := scanNotificationViewFromRows(rows)
		if err != nil {
			return nil, 0, apperrors.HandlePgError(err, tableName, nil)
		}
		views = append(views, v)
	}

	return views, total, nil
}

func scanNotificationView(row pgx.Row) (*domain.NotificationView, error) {
	var (
		id         uuid.UUID
		title      string
		body       string
		nType      string
		targetType string
		isActive   bool
		createdAt  time.Time
		updatedAt  time.Time
	)

	err := row.Scan(&id, &title, &body, &nType, &targetType, &isActive, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, map[string]any{"id": id})
	}

	_ = targetType
	_ = isActive
	_ = updatedAt

	return &domain.NotificationView{
		ID:        id,
		UserID:    uuid.Nil,
		Title:     title,
		Message:   body,
		Type:      nType,
		ReadAt:    nil,
		CreatedAt: createdAt,
	}, nil
}

func scanNotificationViewFromRows(rows pgx.Rows) (*domain.NotificationView, error) {
	var (
		id         uuid.UUID
		title      string
		body       string
		nType      string
		targetType string
		isActive   bool
		createdAt  time.Time
		updatedAt  time.Time
	)

	err := rows.Scan(&id, &title, &body, &nType, &targetType, &isActive, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	_ = targetType
	_ = isActive
	_ = updatedAt

	return &domain.NotificationView{
		ID:        id,
		UserID:    uuid.Nil,
		Title:     title,
		Message:   body,
		Type:      nType,
		ReadAt:    nil,
		CreatedAt: createdAt,
	}, nil
}
