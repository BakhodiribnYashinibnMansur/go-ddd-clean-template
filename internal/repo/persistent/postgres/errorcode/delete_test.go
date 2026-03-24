package errorcode

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

	mockPool.ExpectExec("DELETE FROM error_code").
		WithArgs("ERR_001").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	err = repo.Delete(ctx, "ERR_001")

	require.NoError(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Delete_Error(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	mockPool.ExpectExec("DELETE FROM error_code").
		WithArgs("ERR_FAIL").
		WillReturnError(errors.New("delete failed"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	err = repo.Delete(ctx, "ERR_FAIL")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
