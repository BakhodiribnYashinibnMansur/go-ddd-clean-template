package relation

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

func TestRepo_AddUser(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	relationID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	userID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name          string
		relationID    uuid.UUID
		userID        uuid.UUID
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name:       "success",
			relationID: relationID,
			userID:     userID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO user_relation").
					WithArgs(
						pgxmock.AnyArg(), // user_id
						pgxmock.AnyArg(), // relation_id
						pgxmock.AnyArg(), // created_at
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: false,
		},
		{
			name:       "database error",
			relationID: relationID,
			userID:     userID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO user_relation").
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
			name:       "duplicate user-relation pair",
			relationID: relationID,
			userID:     userID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO user_relation").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnError(errors.New("duplicate key value violates unique constraint"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name:       "foreign key violation - user not found",
			relationID: relationID,
			userID:     uuid.New(),
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO user_relation").
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
			name:       "foreign key violation - relation not found",
			relationID: uuid.New(),
			userID:     userID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO user_relation").
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
			name:       "connection timeout",
			relationID: relationID,
			userID:     userID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO user_relation").
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			tt.setupMock(mockPool)

			repo := &Repo{
				pool:    mockPool,
				builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
				logger:  logger.New("debug"),
			}

			err = repo.AddUser(ctx, tt.relationID, tt.userID)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
