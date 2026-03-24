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

func (r *testRepo) Gets(ctx context.Context, filter *domain.SiteSettingsFilter) ([]*domain.SiteSetting, int, error) {
	query := r.builder.Select(
		"id", "key", "value", "value_type", "category", "description", "is_public", "created_at", "updated_at",
	).From(tableName)

	countQuery := r.builder.Select("COUNT(*)").From(tableName)

	if filter.Key != nil {
		query = query.Where(squirrel.Eq{"key": filter.Key})
		countQuery = countQuery.Where(squirrel.Eq{"key": filter.Key})
	}
	if filter.Category != nil {
		query = query.Where(squirrel.Eq{"category": filter.Category})
		countQuery = countQuery.Where(squirrel.Eq{"category": filter.Category})
	}
	if filter.IsPublic != nil {
		query = query.Where(squirrel.Eq{"is_public": filter.IsPublic})
		countQuery = countQuery.Where(squirrel.Eq{"is_public": filter.IsPublic})
	}

	sql, args, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, err
	}
	var count int
	if err := r.pool.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		return nil, 0, err
	}

	if filter.Pagination != nil {
		if filter.Pagination.Limit > 0 {
			query = query.Limit(uint64(filter.Pagination.Limit))
		}
		if filter.Pagination.Offset > 0 {
			query = query.Offset(uint64(filter.Pagination.Offset))
		}
	}

	query = query.OrderBy("category ASC", "key ASC")
	sql, args, err = query.ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var settings []*domain.SiteSetting
	for rows.Next() {
		var s domain.SiteSetting
		if err := rows.Scan(&s.ID, &s.Key, &s.Value, &s.ValueType, &s.Category, &s.Description, &s.IsPublic, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		settings = append(settings, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return settings, count, nil
}

func TestRepo_Gets_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	category := "general"
	filter := &domain.SiteSettingsFilter{
		SiteSettingFilter: domain.SiteSettingFilter{
			Category: &category,
		},
	}

	mockPool.ExpectQuery("SELECT COUNT").
		WithArgs(category).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(2))

	now := time.Now()
	id1, id2 := uuid.New(), uuid.New()
	listRows := pgxmock.NewRows([]string{
		"id", "key", "value", "value_type", "category", "description", "is_public", "created_at", "updated_at",
	}).
		AddRow(id1, "site_logo", "/logo.png", "string", "general", "Logo", true, now, now).
		AddRow(id2, "site_name", "My Site", "string", "general", "Name", true, now, now)

	mockPool.ExpectQuery("SELECT (.+) FROM " + tableName).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(listRows)

	repo := newTestRepo(mockPool)
	results, count, err := repo.Gets(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Len(t, results, 2)
	assert.Equal(t, "site_logo", results[0].Key)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Gets_Empty(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	filter := &domain.SiteSettingsFilter{}

	mockPool.ExpectQuery("SELECT COUNT").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))

	mockPool.ExpectQuery("SELECT (.+) FROM " + tableName).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "key", "value", "value_type", "category", "description", "is_public", "created_at", "updated_at",
		}))

	repo := newTestRepo(mockPool)
	results, count, err := repo.Gets(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, results)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Gets_CountError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	filter := &domain.SiteSettingsFilter{}

	mockPool.ExpectQuery("SELECT COUNT").
		WillReturnError(errors.New("count failed"))

	repo := newTestRepo(mockPool)
	results, count, err := repo.Gets(ctx, filter)

	require.Error(t, err)
	assert.Equal(t, 0, count)
	assert.Nil(t, results)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Gets_QueryError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	filter := &domain.SiteSettingsFilter{}

	mockPool.ExpectQuery("SELECT COUNT").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(5))

	mockPool.ExpectQuery("SELECT (.+) FROM " + tableName).
		WillReturnError(errors.New("query failed"))

	repo := newTestRepo(mockPool)
	results, count, err := repo.Gets(ctx, filter)

	require.Error(t, err)
	assert.Equal(t, 0, count)
	assert.Nil(t, results)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
