package webhook

import (
	"errors"
	"testing"

	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Delete_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	id := uuid.New()

	// SET deleted_at = NOW() WHERE id = $1
	// squirrel converts uuid to string; use AnyArg
	mockPool.ExpectExec("UPDATE webhooks SET").
		WithArgs(pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	err = repo.Delete(ctx, id)

	require.NoError(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Delete_Error(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	id := uuid.New()

	mockPool.ExpectExec("UPDATE webhooks SET").
		WithArgs(pgxmock.AnyArg()).
		WillReturnError(errors.New("delete failed"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	err = repo.Delete(ctx, id)

	require.Error(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
