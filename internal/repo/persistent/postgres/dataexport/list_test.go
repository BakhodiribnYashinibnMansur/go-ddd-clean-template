package dataexport

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_List_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	filter := domain.DataExportFilter{
		Limit:  10,
		Offset: 0,
	}

	mockPool.ExpectQuery("SELECT COUNT").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(2)))

	now := time.Now()
	userID := "user-1"
	filters := json.RawMessage(`{}`)

	listRows := pgxmock.NewRows([]string{
		"id", "type", "status", "file_url", "filters", "created_by", "created_at", "completed_at",
	}).
		AddRow("exp-1", "users", "completed", nil, filters, &userID, now, &now).
		AddRow("exp-2", "orders", "completed", nil, filters, &userID, now, &now)

	mockPool.ExpectQuery("SELECT (.+) FROM " + table).
		WillReturnRows(listRows)

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	items, total, err := repo.List(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)
	assert.Equal(t, "exp-1", items[0].ID)
	assert.Equal(t, "exp-2", items[1].ID)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_List_Empty(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	filter := domain.DataExportFilter{Limit: 10}

	mockPool.ExpectQuery("SELECT COUNT").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mockPool.ExpectQuery("SELECT (.+) FROM " + table).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "type", "status", "file_url", "filters", "created_by", "created_at", "completed_at",
		}))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	items, total, err := repo.List(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, items)
	assert.NotNil(t, items)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_List_WithTypeFilter(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	filter := domain.DataExportFilter{
		Type:  "users",
		Limit: 10,
	}

	mockPool.ExpectQuery("SELECT COUNT").
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mockPool.ExpectQuery("SELECT (.+) FROM " + table).
		WithArgs(pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "type", "status", "file_url", "filters", "created_by", "created_at", "completed_at",
		}))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	items, _, err := repo.List(ctx, filter)

	require.NoError(t, err)
	assert.NotNil(t, items)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_List_CountError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	filter := domain.DataExportFilter{Limit: 10}

	mockPool.ExpectQuery("SELECT COUNT").
		WillReturnError(errors.New("count failed"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	items, total, err := repo.List(ctx, filter)

	require.Error(t, err)
	assert.Equal(t, int64(0), total)
	assert.Nil(t, items)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_List_QueryError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	filter := domain.DataExportFilter{Limit: 10}

	mockPool.ExpectQuery("SELECT COUNT").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(5)))

	mockPool.ExpectQuery("SELECT (.+) FROM " + table).
		WillReturnError(errors.New("query failed"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	items, total, err := repo.List(ctx, filter)

	require.Error(t, err)
	assert.Equal(t, int64(0), total)
	assert.Nil(t, items)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
