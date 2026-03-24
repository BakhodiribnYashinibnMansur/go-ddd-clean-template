package setting

import (
	"errors"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Delete(t *testing.T) {
	ctx := t.Context()
	userID := uuid.New()

	tests := []struct {
		name          string
		userID        uuid.UUID
		key           string
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
	}{
		{
			name:   "success",
			userID: userID,
			key:    "theme",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM user_settings").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			expectedError: false,
		},
		{
			name:   "success - no rows affected",
			userID: userID,
			key:    "nonexistent",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM user_settings").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			expectedError: false,
		},
		{
			name:   "database error",
			userID: userID,
			key:    "theme",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM user_settings").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(errors.New("connection refused"))
			},
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

			err = repo.Delete(ctx, tt.userID, tt.key)
			if tt.expectedError {
				require.Error(t, err)
			} else {
				// HandlePgError returns *AppError; a nil *AppError wrapped in
				// an error interface is non-nil per Go rules, so use assert.Nil
				// which checks the underlying value via reflect.
				assert.Nil(t, err)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
