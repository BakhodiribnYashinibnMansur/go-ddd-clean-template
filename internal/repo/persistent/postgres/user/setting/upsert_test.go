package setting

import (
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

func TestRepo_Upsert(t *testing.T) {
	ctx := t.Context()
	userID := uuid.New()
	now := time.Now()

	tests := []struct {
		name          string
		setting       *domain.UserSetting
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
	}{
		{
			name: "success - insert new setting",
			setting: &domain.UserSetting{
				ID: uuid.New(), UserID: userID, Key: "theme", Value: "dark",
				CreatedAt: now, UpdatedAt: now,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO user_settings").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // user_id
						pgxmock.AnyArg(), // key
						pgxmock.AnyArg(), // value
						pgxmock.AnyArg(), // created_at
						pgxmock.AnyArg(), // updated_at
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: false,
		},
		{
			name: "success - upsert (conflict update)",
			setting: &domain.UserSetting{
				ID: uuid.New(), UserID: userID, Key: "theme", Value: "light",
				CreatedAt: now, UpdatedAt: now,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO user_settings").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: false,
		},
		{
			name: "database error",
			setting: &domain.UserSetting{
				ID: uuid.New(), UserID: userID, Key: "theme", Value: "dark",
				CreatedAt: now, UpdatedAt: now,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO user_settings").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnError(errors.New("connection refused"))
			},
			expectedError: true,
		},
		{
			name: "foreign key violation",
			setting: &domain.UserSetting{
				ID: uuid.New(), UserID: uuid.New(), Key: "theme", Value: "dark",
				CreatedAt: now, UpdatedAt: now,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO user_settings").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnError(errors.New("foreign key constraint"))
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

			err = repo.Upsert(ctx, tt.setting)
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
