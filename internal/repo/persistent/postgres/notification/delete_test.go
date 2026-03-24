package notification

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

func TestRepo_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		id            uuid.UUID
		setupMock     func(pgxmock.PgxPoolIface, uuid.UUID)
		expectedError bool
	}{
		{
			name: "success",
			id:   uuid.New(),
			setupMock: func(mock pgxmock.PgxPoolIface, id uuid.UUID) {
				mock.ExpectExec("DELETE FROM notifications").
					WithArgs(pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			expectedError: false,
		},
		{
			name: "database error",
			id:   uuid.New(),
			setupMock: func(mock pgxmock.PgxPoolIface, id uuid.UUID) {
				mock.ExpectExec("DELETE FROM notifications").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("delete failed"))
			},
			expectedError: true,
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

			err = repo.Delete(t.Context(), tt.id)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
