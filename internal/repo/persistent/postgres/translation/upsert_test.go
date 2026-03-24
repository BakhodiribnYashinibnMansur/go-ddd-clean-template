package translation

import (
	"errors"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Upsert_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	entityType := "role"
	entityID := uuid.New()
	langCode := "uz"
	data := map[string]string{"title": "Sarlavha"}

	mockPool.ExpectExec("INSERT INTO " + tableName).
		WithArgs(entityType, entityID, langCode, pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	err = repo.Upsert(ctx, entityType, entityID, langCode, data)

	require.NoError(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Upsert_DBError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	entityType := "role"
	entityID := uuid.New()
	langCode := "uz"
	data := map[string]string{"title": "Sarlavha"}

	mockPool.ExpectExec("INSERT INTO " + tableName).
		WithArgs(entityType, entityID, langCode, pgxmock.AnyArg()).
		WillReturnError(errors.New("connection refused"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	err = repo.Upsert(ctx, entityType, entityID, langCode, data)

	require.Error(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
