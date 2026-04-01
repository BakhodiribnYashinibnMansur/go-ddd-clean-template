package postgres

import (
	"context"
	"time"

	"gct/internal/notification/domain"
	"gct/internal/shared/domain/consts"
	apperrors "gct/internal/shared/infrastructure/errors"
	"gct/internal/shared/infrastructure/pgxutil"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableNotifications

var writeColumns = []string{
	"id", "title", "body", "type", "target_type", "is_active", "created_at", "updated_at",
}

// NotificationWriteRepo implements domain.NotificationRepository using PostgreSQL.
type NotificationWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewNotificationWriteRepo creates a new NotificationWriteRepo.
func NewNotificationWriteRepo(pool *pgxpool.Pool) *NotificationWriteRepo {
	return &NotificationWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Save inserts a new Notification aggregate into the database.
func (r *NotificationWriteRepo) Save(ctx context.Context, n *domain.Notification) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "NotificationWriteRepo.Save")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			n.ID(),
			n.Title(),
			n.Message(),
			n.Type(),
			"all",
			true,
			n.CreatedAt(),
			n.CreatedAt(),
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

// FindByID retrieves a Notification aggregate by ID.
func (r *NotificationWriteRepo) FindByID(ctx context.Context, id uuid.UUID) (result *domain.Notification, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "NotificationWriteRepo.FindByID")
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
	return scanNotification(row)
}

// Update updates a Notification aggregate in the database.
func (r *NotificationWriteRepo) Update(ctx context.Context, n *domain.Notification) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "NotificationWriteRepo.Update")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Update(tableName).
		Set("is_active", false).
		Where(squirrel.Eq{"id": n.ID()}).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildUpdate)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// Delete removes a Notification by ID.
func (r *NotificationWriteRepo) Delete(ctx context.Context, id uuid.UUID) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "NotificationWriteRepo.Delete")
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

func scanNotification(row pgx.Row) (*domain.Notification, error) {
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

	return domain.ReconstructNotification(id, createdAt, uuid.Nil, title, body, nType, nil), nil
}
