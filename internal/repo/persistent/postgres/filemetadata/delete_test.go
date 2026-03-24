package filemetadata

import (
	"errors"
	"testing"

	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Delete_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	id := "file-123"

	mockPool.ExpectExec("DELETE FROM " + table).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	err = repo.Delete(ctx, id)

	require.NoError(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Delete_DBError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	id := "file-456"

	mockPool.ExpectExec("DELETE FROM " + table).
		WithArgs(id).
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
