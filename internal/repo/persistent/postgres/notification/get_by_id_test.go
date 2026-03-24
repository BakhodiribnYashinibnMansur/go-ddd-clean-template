package notification

import (
	"errors"
	"testing"
	"time"

	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_GetByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		id            uuid.UUID
		setupMock     func(pgxmock.PgxPoolIface, uuid.UUID)
		expectedError bool
		checkResult   bool
	}{
		{
			name: "success",
			id:   uuid.New(),
			setupMock: func(mock pgxmock.PgxPoolIface, id uuid.UUID) {
				now := time.Now()
				mock.ExpectQuery("SELECT (.+) FROM notifications").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(
						pgxmock.NewRows([]string{
							"id", "title", "body", "type", "target_type",
							"is_active", "created_at", "updated_at",
						}).AddRow(id, "Welcome", "Hello!", "info", "all", true, now, now),
					)
			},
			expectedError: false,
			checkResult:   true,
		},
		{
			name: "not found",
			id:   uuid.New(),
			setupMock: func(mock pgxmock.PgxPoolIface, id uuid.UUID) {
				mock.ExpectQuery("SELECT (.+) FROM notifications").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("no rows in result set"))
			},
			expectedError: true,
			checkResult:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			tt.setupMock(mockPool, tt.id)

			repo := &Repo{
				pool:    mockPool,
				builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
				logger:  logger.New("debug"),
			}

			result, err := repo.GetByID(t.Context(), tt.id)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.id, result.ID)
				assert.Equal(t, "Welcome", result.Title)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
