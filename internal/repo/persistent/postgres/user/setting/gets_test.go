package setting

import (
	"errors"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Gets(t *testing.T) {
	ctx := t.Context()
	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name          string
		userID        uuid.UUID
		setupMock     func(pgxmock.PgxPoolIface)
		expectedCount int
		expectedError bool
	}{
		{
			name:   "success - returns settings",
			userID: userID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "user_id", "key", "value", "created_at", "updated_at"}).
					AddRow(uuid.New(), userID, "theme", "dark", now, now).
					AddRow(uuid.New(), userID, "lang", "en", now, now)
				mock.ExpectQuery("SELECT (.+) FROM user_settings").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name:   "success - empty result",
			userID: userID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "user_id", "key", "value", "created_at", "updated_at"})
				mock.ExpectQuery("SELECT (.+) FROM user_settings").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedCount: 0,
			expectedError: false,
		},
		{
			name:   "query error",
			userID: userID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM user_settings").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("connection refused"))
			},
			expectedCount: 0,
			expectedError: true,
		},
		{
			name:   "scan error",
			userID: userID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "user_id", "key", "value", "created_at", "updated_at"}).
					AddRow("not-a-uuid", userID, "theme", "dark", now, now)
				mock.ExpectQuery("SELECT (.+) FROM user_settings").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedCount: 0,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			tt.setupMock(mockPool)

			repo := &Repo{
				pool:    mockPool,
				builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
			}

			settings, err := repo.Gets(ctx, tt.userID)
			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, settings, tt.expectedCount)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
