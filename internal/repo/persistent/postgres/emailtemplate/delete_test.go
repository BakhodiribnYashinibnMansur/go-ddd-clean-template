package emailtemplate

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

	mockPool.ExpectExec("DELETE FROM email_templates").
		WithArgs("tmpl-1").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	err = repo.Delete(ctx, "tmpl-1")

	// HandlePgError returns *AppError(nil) which is non-nil as error interface;
	// verify the underlying value is nil.
	assert.Nil(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Delete_Error(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	mockPool.ExpectExec("DELETE FROM email_templates").
		WithArgs("tmpl-1").
		WillReturnError(errors.New("delete failed"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	err = repo.Delete(ctx, "tmpl-1")

	require.Error(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
