package translation

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"gct/internal/domain"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Gets_Success(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	entityID := uuid.New()
	filter := domain.TranslationFilter{
		EntityType: "role",
		EntityID:   entityID,
	}

	id1 := uuid.New()
	now := time.Now()
	dataJSON, _ := json.Marshal(map[string]string{"title": "Sarlavha"})

	rows := pgxmock.NewRows([]string{"id", "entity_type", "entity_id", "lang_code", "data", "created_at", "updated_at"}).
		AddRow(id1, "role", entityID, "uz", dataJSON, now, now)

	mockPool.ExpectQuery("SELECT (.+) FROM " + tableName).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(rows)

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	result, err := repo.Gets(ctx, filter)

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "uz", result[0].LangCode)
	assert.Equal(t, "Sarlavha", result[0].Data["title"])
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Gets_WithLangCodeFilter(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	entityID := uuid.New()
	lang := "en"
	filter := domain.TranslationFilter{
		EntityType: "role",
		EntityID:   entityID,
		LangCode:   &lang,
	}

	id1 := uuid.New()
	now := time.Now()
	dataJSON, _ := json.Marshal(map[string]string{"title": "Title"})

	rows := pgxmock.NewRows([]string{"id", "entity_type", "entity_id", "lang_code", "data", "created_at", "updated_at"}).
		AddRow(id1, "role", entityID, "en", dataJSON, now, now)

	mockPool.ExpectQuery("SELECT (.+) FROM " + tableName).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(rows)

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	result, err := repo.Gets(ctx, filter)

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "en", result[0].LangCode)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Gets_Empty(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	entityID := uuid.New()
	filter := domain.TranslationFilter{
		EntityType: "role",
		EntityID:   entityID,
	}

	rows := pgxmock.NewRows([]string{"id", "entity_type", "entity_id", "lang_code", "data", "created_at", "updated_at"})

	mockPool.ExpectQuery("SELECT (.+) FROM " + tableName).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnRows(rows)

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	result, err := repo.Gets(ctx, filter)

	require.NoError(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}

func TestRepo_Gets_QueryError(t *testing.T) {
	ctx := t.Context()

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	entityID := uuid.New()
	filter := domain.TranslationFilter{
		EntityType: "role",
		EntityID:   entityID,
	}

	mockPool.ExpectQuery("SELECT (.+) FROM " + tableName).
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnError(errors.New("query failed"))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	result, err := repo.Gets(ctx, filter)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
