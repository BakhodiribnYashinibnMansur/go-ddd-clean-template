package postgres

import (
	"context"
	"time"

	"gct/internal/platform/domain/consts"
	apperrors "gct/internal/platform/infrastructure/errors"
	"gct/internal/platform/infrastructure/pgxutil"
	"gct/internal/context/iam/usersetting/domain"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const tableName = consts.TableUserSettings

var writeColumns = []string{
	"id", "user_id", "key", "value", "created_at", "updated_at",
}

// UserSettingWriteRepo implements domain.UserSettingRepository using PostgreSQL.
type UserSettingWriteRepo struct {
	pool    *pgxpool.Pool
	builder squirrel.StatementBuilderType
}

// NewUserSettingWriteRepo creates a new UserSettingWriteRepo.
func NewUserSettingWriteRepo(pool *pgxpool.Pool) *UserSettingWriteRepo {
	return &UserSettingWriteRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Upsert inserts or updates a UserSetting aggregate in the database.
func (r *UserSettingWriteRepo) Upsert(ctx context.Context, us *domain.UserSetting) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "UserSettingWriteRepo.Upsert")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Insert(tableName).
		Columns(writeColumns...).
		Values(
			us.ID(),
			us.UserID(),
			us.Key(),
			us.Value(),
			us.CreatedAt(),
			us.UpdatedAt(),
		).
		Suffix("ON CONFLICT (user_id, key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at").
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	if _, err = r.pool.Exec(ctx, sql, args...); err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

// FindByUserIDAndKey retrieves a UserSetting aggregate by user ID and key.
func (r *UserSettingWriteRepo) FindByUserIDAndKey(ctx context.Context, userID uuid.UUID, key string) (result *domain.UserSetting, err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "UserSettingWriteRepo.FindByUserIDAndKey")
	defer func() { end(err) }()

	sql, args, err := r.builder.
		Select(writeColumns...).
		From(tableName).
		Where(squirrel.Eq{"user_id": userID, "key": key}).
		ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildQuery)
	}

	row := r.pool.QueryRow(ctx, sql, args...)
	return scanUserSetting(row)
}

// Delete removes a user setting by its ID.
func (r *UserSettingWriteRepo) Delete(ctx context.Context, id uuid.UUID) (err error) {
	ctx, end := pgxutil.RepoSpan(ctx, "UserSettingWriteRepo.Delete")
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

func scanUserSetting(row pgx.Row) (*domain.UserSetting, error) {
	var (
		id        uuid.UUID
		userID    uuid.UUID
		key       string
		value     string
		createdAt time.Time
		updatedAt time.Time
	)

	err := row.Scan(&id, &userID, &key, &value, &createdAt, &updatedAt)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return domain.ReconstructUserSetting(id, createdAt, updatedAt, userID, key, value), nil
}
