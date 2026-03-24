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

func TestRepo_Create_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	input := CreateErrorCodeInput{
		Code:       "ERR_001",
		Message:    "Something went wrong",
		HTTPStatus: 500,
		Category:   domain.CategorySystem,
		Severity:   domain.SeverityHigh,
		Retryable:  true,
		RetryAfter: 30,
		Suggestion: "Try again later",
	}

	id := uuid.New()
	now := time.Now()
	rows := pgxmock.NewRows([]string{
		"id", "code", "message", "http_status", "category", "severity",
		"retryable", "retry_after", "suggestion", "created_at", "updated_at",
	}).AddRow(
		id, input.Code, input.Message, input.HTTPStatus, input.Category, input.Severity,
		input.Retryable, input.RetryAfter, input.Suggestion, now, now,
	)

	mockPool.ExpectQuery("INSERT INTO error_code").
		WithArgs(
			input.Code,
			input.Message,
			input.HTTPStatus,
			input.Category,
			input.Severity,
			input.Retryable,
			input.RetryAfter,
			input.Suggestion,
		).
		WillReturnRows(rows)

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	result, err := repo.Create(ctx, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, id, result.ID)
	assert.Equal(t, input.Code, result.Code)
	assert.Equal(t, input.Message, result.Message)
	assert.Equal(t, input.HTTPStatus, result.HTTPStatus)
	assert.Equal(t, input.Category, result.Category)
	assert.Equal(t, input.Severity, result.Severity)
	assert.Equal(t, input.Retryable, result.Retryable)
	assert.Equal(t, input.RetryAfter, result.RetryAfter)
	assert.Equal(t, input.Suggestion, result.Suggestion)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Create_DBError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	input := CreateErrorCodeInput{
		Code:       "ERR_DUP",
		Message:    "Duplicate",
		HTTPStatus: 409,
		Category:   domain.CategoryValidation,
		Severity:   domain.SeverityLow,
	}

	mockPool.ExpectQuery("INSERT INTO error_code").
		WithArgs(
			input.Code,
			input.Message,
			input.HTTPStatus,
			input.Category,
			input.Severity,
			input.Retryable,
			input.RetryAfter,
			input.Suggestion,
		).
		WillReturnError(errors.New("unique violation"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	result, err := repo.Create(ctx, input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unique violation")
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
