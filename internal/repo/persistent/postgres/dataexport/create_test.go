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

func TestRepo_Create_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	now := time.Now()
	userID := "user-1"
	e := &domain.DataExport{
		ID:          "exp-1",
		Type:        "users",
		Status:      "completed",
		Filters:     json.RawMessage(`{"status":"active"}`),
		CreatedBy:   &userID,
		CompletedAt: &now,
	}

	createdAt := time.Now()
	rows := pgxmock.NewRows([]string{"created_at"}).AddRow(createdAt)

	mockPool.ExpectQuery("INSERT INTO " + table).
		WithArgs(
			e.ID, e.Type, e.Status, e.FileURL,
			pgxmock.AnyArg(), // filters as string
			e.CreatedBy, e.CompletedAt,
		).
		WillReturnRows(rows)

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	err = repo.Create(ctx, e)

	require.NoError(t, err)
	assert.Equal(t, createdAt, e.CreatedAt)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Create_DBError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	userID := "user-1"
	now := time.Now()
	e := &domain.DataExport{
		ID:          "exp-2",
		Type:        "orders",
		Status:      "completed",
		Filters:     json.RawMessage(`{}`),
		CreatedBy:   &userID,
		CompletedAt: &now,
	}

	mockPool.ExpectQuery("INSERT INTO " + table).
		WithArgs(
			e.ID, e.Type, e.Status, e.FileURL,
			pgxmock.AnyArg(),
			e.CreatedBy, e.CompletedAt,
		).
		WillReturnError(errors.New("connection refused"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	err = repo.Create(ctx, e)

	require.Error(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
