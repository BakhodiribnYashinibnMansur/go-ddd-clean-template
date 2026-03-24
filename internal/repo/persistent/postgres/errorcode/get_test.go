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

func TestRepo_GetByCode_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	id := uuid.New()
	now := time.Now()
	rows := pgxmock.NewRows([]string{
		"id", "code", "message", "http_status", "category", "severity",
		"retryable", "retry_after", "suggestion", "created_at", "updated_at",
	}).AddRow(
		id, "ERR_001", "Something went wrong", 500,
		domain.CategorySystem, domain.SeverityHigh,
		true, 30, "Try again later", now, now,
	)

	mockPool.ExpectQuery("SELECT (.+) FROM error_code").
		WithArgs("ERR_001").
		WillReturnRows(rows)

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	result, err := repo.GetByCode(ctx, "ERR_001")

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, id, result.ID)
	assert.Equal(t, "ERR_001", result.Code)
	assert.Equal(t, "Something went wrong", result.Message)
	assert.Equal(t, 500, result.HTTPStatus)
	assert.Equal(t, domain.CategorySystem, result.Category)
	assert.Equal(t, domain.SeverityHigh, result.Severity)
	assert.True(t, result.Retryable)
	assert.Equal(t, 30, result.RetryAfter)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_GetByCode_NotFound(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	mockPool.ExpectQuery("SELECT (.+) FROM error_code").
		WithArgs("NONEXISTENT").
		WillReturnError(errors.New("no rows in result set"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	result, err := repo.GetByCode(ctx, "NONEXISTENT")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_GetByCode_DBError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	mockPool.ExpectQuery("SELECT (.+) FROM error_code").
		WithArgs("ERR_002").
		WillReturnError(errors.New("connection refused"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	result, err := repo.GetByCode(ctx, "ERR_002")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "connection refused")
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
