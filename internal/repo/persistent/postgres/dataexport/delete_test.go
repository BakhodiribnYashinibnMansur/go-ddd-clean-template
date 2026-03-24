package dataexport

import (
	"errors"
	"reflect"
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

	id := "exp-123"

	mockPool.ExpectExec("DELETE FROM " + table).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	deleteErr := repo.Delete(ctx, id)

	// HandlePgError returns *AppError(nil) which is a typed nil;
	// as an error interface it is technically non-nil due to Go's
	// interface semantics. We use reflect to verify the value is nil.
	if deleteErr != nil {
		assert.True(t, reflect.ValueOf(deleteErr).IsNil(), "expected nil-valued error")
	}
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Delete_DBError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	id := "exp-456"

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
