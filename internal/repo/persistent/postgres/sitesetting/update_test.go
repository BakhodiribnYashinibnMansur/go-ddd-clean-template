package sitesetting

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/domain"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (r *testRepo) Update(ctx context.Context, setting *domain.SiteSetting) error {
	setting.UpdatedAt = time.Now()

	sql, args, err := r.builder.
		Update(tableName).
		Set("value", setting.Value).
		Set("value_type", setting.ValueType).
		Set("category", setting.Category).
		Set("description", setting.Description).
		Set("is_public", setting.IsPublic).
		Set("updated_at", setting.UpdatedAt).
		Where(squirrel.Eq{"id": setting.ID}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func (r *testRepo) UpdateByKey(ctx context.Context, key, value string) error {
	sql, args, err := r.builder.
		Update(tableName).
		Set("value", value).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"key": key}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	return err
}

func TestRepo_Update_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	setting := &domain.SiteSetting{
		ID:          uuid.New(),
		Key:         "site_name",
		Value:       "New Name",
		ValueType:   "string",
		Category:    "general",
		Description: "Site name",
		IsPublic:    true,
	}

	mockPool.ExpectExec("UPDATE " + tableName + " SET").
		WithArgs(
			setting.Value, setting.ValueType, setting.Category,
			setting.Description, setting.IsPublic,
			pgxmock.AnyArg(), // updated_at
			pgxmock.AnyArg(), // id (uuid -> string by squirrel)
		).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	repo := newTestRepo(mockPool)
	err = repo.Update(ctx, setting)

	require.NoError(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Update_DBError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	setting := &domain.SiteSetting{
		ID:    uuid.New(),
		Key:   "site_name",
		Value: "Fail",
	}

	mockPool.ExpectExec("UPDATE " + tableName + " SET").
		WithArgs(
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(),
			pgxmock.AnyArg(), // id (uuid -> string by squirrel)
		).
		WillReturnError(errors.New("update failed"))

	repo := newTestRepo(mockPool)
	err = repo.Update(ctx, setting)

	require.Error(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_UpdateByKey_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	mockPool.ExpectExec("UPDATE " + tableName + " SET").
		WithArgs("new_value", pgxmock.AnyArg(), "site_name"). // value, updated_at, key
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	repo := newTestRepo(mockPool)
	err = repo.UpdateByKey(ctx, "site_name", "new_value")

	require.NoError(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_UpdateByKey_DBError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	mockPool.ExpectExec("UPDATE " + tableName + " SET").
		WithArgs("value", pgxmock.AnyArg(), "missing_key").
		WillReturnError(errors.New("key not found"))

	repo := newTestRepo(mockPool)
	err = repo.UpdateByKey(ctx, "missing_key", "value")

	require.Error(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
