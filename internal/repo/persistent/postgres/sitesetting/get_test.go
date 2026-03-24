package sitesetting

import (
	"context"
	"errors"
	"testing"
	"time"

	"gct/internal/domain"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testPool wraps pgxmock to satisfy *pgxpool.Pool usage patterns.
// Since Repo.pool is *pgxpool.Pool (concrete), we introduce a local
// testRepo that mirrors Repo but accepts the pgxmock interface.
type testPool interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type testRepo struct {
	pool    testPool
	builder squirrel.StatementBuilderType
}

func (r *testRepo) Get(ctx context.Context, filter *domain.SiteSettingFilter) (*domain.SiteSetting, error) {
	query := r.builder.Select(
		"id", "key", "value", "value_type", "category", "description", "is_public", "created_at", "updated_at",
	).From(tableName)

	if filter.ID != nil {
		query = query.Where(squirrel.Eq{"id": filter.ID})
	}
	if filter.Key != nil {
		query = query.Where(squirrel.Eq{"key": filter.Key})
	}
	if filter.Category != nil {
		query = query.Where(squirrel.Eq{"category": filter.Category})
	}
	if filter.IsPublic != nil {
		query = query.Where(squirrel.Eq{"is_public": filter.IsPublic})
	}
	query = query.Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var setting domain.SiteSetting
	err = r.pool.QueryRow(ctx, sql, args...).Scan(
		&setting.ID, &setting.Key, &setting.Value, &setting.ValueType,
		&setting.Category, &setting.Description, &setting.IsPublic,
		&setting.CreatedAt, &setting.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func newTestRepo(pool testPool) *testRepo {
	return &testRepo{
		pool:    pool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func TestRepo_Get_ByID(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	id := uuid.New()
	now := time.Now()

	rows := pgxmock.NewRows([]string{
		"id", "key", "value", "value_type", "category", "description", "is_public", "created_at", "updated_at",
	}).AddRow(id, "site_name", "My Site", "string", "general", "Site name", true, now, now)

	mockPool.ExpectQuery("SELECT (.+) FROM " + tableName).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(rows)

	repo := newTestRepo(mockPool)
	filter := &domain.SiteSettingFilter{ID: &id}

	result, err := repo.Get(ctx, filter)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, id, result.ID)
	assert.Equal(t, "site_name", result.Key)
	assert.Equal(t, "My Site", result.Value)
	assert.True(t, result.IsPublic)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Get_ByKey(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	id := uuid.New()
	now := time.Now()
	key := "maintenance_mode"

	rows := pgxmock.NewRows([]string{
		"id", "key", "value", "value_type", "category", "description", "is_public", "created_at", "updated_at",
	}).AddRow(id, key, "false", "boolean", "maintenance", "Enable maintenance", false, now, now)

	mockPool.ExpectQuery("SELECT (.+) FROM " + tableName).
		WithArgs(key).
		WillReturnRows(rows)

	repo := newTestRepo(mockPool)
	filter := &domain.SiteSettingFilter{Key: &key}

	result, err := repo.Get(ctx, filter)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, key, result.Key)
	assert.Equal(t, "false", result.Value)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Get_Error(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	id := uuid.New()

	mockPool.ExpectQuery("SELECT (.+) FROM " + tableName).
		WithArgs(pgxmock.AnyArg()).
		WillReturnError(errors.New("no rows"))

	repo := newTestRepo(mockPool)
	filter := &domain.SiteSettingFilter{ID: &id}

	result, err := repo.Get(ctx, filter)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
