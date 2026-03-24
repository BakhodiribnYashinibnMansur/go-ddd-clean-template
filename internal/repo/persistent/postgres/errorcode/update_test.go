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

func TestRepo_Update_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	newMessage := "Updated message"
	newStatus := 503
	input := UpdateErrorCodeInput{
		Message:    &newMessage,
		HTTPStatus: &newStatus,
	}

	id := uuid.New()
	now := time.Now()
	rows := pgxmock.NewRows([]string{
		"id", "code", "message", "http_status", "category", "severity",
		"retryable", "retry_after", "suggestion", "created_at", "updated_at",
	}).AddRow(
		id, "ERR_001", newMessage, newStatus,
		domain.CategorySystem, domain.SeverityHigh,
		true, 30, "Try again", now, now,
	)

	mockPool.ExpectQuery("UPDATE error_code SET").
		WithArgs(
			pgxmock.AnyArg(), // updated_at (time.Now())
			newMessage,
			newStatus,
			"ERR_001",
		).
		WillReturnRows(rows)

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	result, err := repo.Update(ctx, "ERR_001", input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, id, result.ID)
	assert.Equal(t, "ERR_001", result.Code)
	assert.Equal(t, newMessage, result.Message)
	assert.Equal(t, newStatus, result.HTTPStatus)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Update_AllFields(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	msg := "Full update"
	status := 502
	cat := domain.CategoryAuth
	sev := domain.SeverityCritical
	retry := true
	retryAfter := 120
	suggestion := "Contact admin"

	input := UpdateErrorCodeInput{
		Message:    &msg,
		HTTPStatus: &status,
		Category:   &cat,
		Severity:   &sev,
		Retryable:  &retry,
		RetryAfter: &retryAfter,
		Suggestion: &suggestion,
	}

	id := uuid.New()
	now := time.Now()
	rows := pgxmock.NewRows([]string{
		"id", "code", "message", "http_status", "category", "severity",
		"retryable", "retry_after", "suggestion", "created_at", "updated_at",
	}).AddRow(
		id, "ERR_003", msg, status, cat, sev, retry, retryAfter, suggestion, now, now,
	)

	mockPool.ExpectQuery("UPDATE error_code SET").
		WithArgs(
			pgxmock.AnyArg(), // updated_at
			msg,
			status,
			cat,
			sev,
			retry,
			retryAfter,
			suggestion,
			"ERR_003",
		).
		WillReturnRows(rows)

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	result, err := repo.Update(ctx, "ERR_003", input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, msg, result.Message)
	assert.Equal(t, status, result.HTTPStatus)
	assert.Equal(t, cat, result.Category)
	assert.Equal(t, sev, result.Severity)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Update_DBError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	msg := "Updated"
	input := UpdateErrorCodeInput{
		Message: &msg,
	}

	mockPool.ExpectQuery("UPDATE error_code SET").
		WithArgs(
			pgxmock.AnyArg(), // updated_at
			msg,
			"ERR_FAIL",
		).
		WillReturnError(errors.New("update failed"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	result, err := repo.Update(ctx, "ERR_FAIL", input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "update failed")
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
