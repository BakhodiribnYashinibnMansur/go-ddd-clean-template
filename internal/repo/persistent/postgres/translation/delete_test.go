package translation

import (
	"errors"
	"testing"

	"gct/internal/domain"

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

	entityID := uuid.New()
	filter := domain.TranslationFilter{
		EntityType: "role",
		EntityID:   entityID,
	}

	mockPool.ExpectExec("DELETE FROM " + tableName).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	err = repo.Delete(ctx, filter)

	require.NoError(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Delete_WithLangCode(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	entityID := uuid.New()
	lang := "uz"
	filter := domain.TranslationFilter{
		EntityType: "role",
		EntityID:   entityID,
		LangCode:   &lang,
	}

	mockPool.ExpectExec("DELETE FROM " + tableName).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	err = repo.Delete(ctx, filter)

	require.NoError(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Delete_DBError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	entityID := uuid.New()
	filter := domain.TranslationFilter{
		EntityType: "role",
		EntityID:   entityID,
	}

	mockPool.ExpectExec("DELETE FROM " + tableName).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnError(errors.New("delete failed"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	err = repo.Delete(ctx, filter)

	require.Error(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
