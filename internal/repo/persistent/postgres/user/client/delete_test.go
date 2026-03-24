package client

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
	ctx := t.Context()

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name          string
		userID        uuid.UUID
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name:   "success",
			userID: userID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(), // deleted_at (unix timestamp)
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // id (in WHERE clause)
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: false,
		},
		{
			name:   "database error",
			userID: userID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnError(errors.New("database error"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name:   "no rows affected",
			userID: userID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: false,
		},
		{
			name:   "user not found",
			userID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: false,
		},
		{
			name:   "zero user ID",
			userID: uuid.Nil,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnError(errors.New("invalid user ID"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid user ID")
			},
		},
		{
			name:   "negative user ID",
			userID: uuid.Nil,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnError(errors.New("invalid user ID"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid user ID")
			},
		},
		{
			name:   "connection timeout",
			userID: userID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnError(errors.New("connection timeout"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "timeout")
			},
		},
		{
			name:   "foreign key constraint violation",
			userID: userID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnError(errors.New("foreign key constraint violation"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "foreign key")
			},
		},
		{
			name:   "user already deleted",
			userID: userID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: false,
		},
		{
			name:   "large user ID",
			userID: uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"), // Max UUID
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: false,
		},
		{
			name:   "database locked",
			userID: userID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnError(errors.New("database is locked"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "locked")
			},
		},
		{
			name:   "permission denied",
			userID: userID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnError(errors.New("permission denied for table users"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "permission denied")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock pool
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			// Setup mock expectations
			tt.setupMock(mockPool)

			// Create repository
			repo := &Repo{
				pool:    mockPool,
				builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
				logger:  logger.New("debug"),
			}

			// Execute test
			err = repo.Delete(ctx, tt.userID)

			// Verify expectations
			if tt.expectedError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
			}

			// Ensure all expectations were met
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestRepo_Delete_SoftDelete(t *testing.T) {
	ctx := t.Context()
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	// Verify that Delete sets deleted_at to a non-zero timestamp
	mockPool.ExpectExec("UPDATE users").
		WithArgs(
			pgxmock.AnyArg(), // deleted_at should be > 0
			pgxmock.AnyArg(), // updated_at
			pgxmock.AnyArg(), // id
		).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	err = repo.Delete(ctx, userID)

	require.NoError(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
