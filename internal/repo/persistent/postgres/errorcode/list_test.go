package errorcode

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_List_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	id1, id2 := uuid.New(), uuid.New()
	now := time.Now()

	rows := pgxmock.NewRows([]string{
		"id", "code", "message", "http_status", "category", "severity",
		"retryable", "retry_after", "suggestion", "created_at", "updated_at",
	}).
		AddRow(id1, "ERR_001", "Error one", 400, domain.CategoryValidation, domain.SeverityLow, false, 0, "Fix input", now, now).
		AddRow(id2, "ERR_002", "Error two", 500, domain.CategorySystem, domain.SeverityHigh, true, 60, "Retry", now, now)

	mockPool.ExpectQuery("SELECT (.+) FROM error_code").
		WillReturnRows(rows)

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	result, err := repo.List(ctx)

	require.NoError(t, err)
	require.Len(t, result, 2)
	assert.Equal(t, "ERR_001", result[0].Code)
	assert.Equal(t, "ERR_002", result[1].Code)
	assert.Equal(t, 400, result[0].HTTPStatus)
	assert.Equal(t, 500, result[1].HTTPStatus)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_List_Empty(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	rows := pgxmock.NewRows([]string{
		"id", "code", "message", "http_status", "category", "severity",
		"retryable", "retry_after", "suggestion", "created_at", "updated_at",
	})

	mockPool.ExpectQuery("SELECT (.+) FROM error_code").
		WillReturnRows(rows)

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	result, err := repo.List(ctx)

	require.NoError(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_List_DBError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	mockPool.ExpectQuery("SELECT (.+) FROM error_code").
		WillReturnError(errors.New("query failed"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	result, err := repo.List(ctx)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "query failed")
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
