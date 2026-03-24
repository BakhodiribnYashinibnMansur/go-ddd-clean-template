package role

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

func TestRepo_AddPermission(t *testing.T) {
	ctx := t.Context()

	roleID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	permID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name          string
		roleID        uuid.UUID
		permID        uuid.UUID
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name:   "success",
			roleID: roleID,
			permID: permID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO role_permission").
					WithArgs(
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // permission_id
						pgxmock.AnyArg(), // created_at
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: false,
		},
		{
			name:   "duplicate error",
			roleID: roleID,
			permID: permID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO role_permission").
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
			name:   "database error",
			roleID: roleID,
			permID: permID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO role_permission").
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
				logger:  logger.New("debug"),
			}

			err = repo.AddPermission(ctx, tt.roleID, tt.permID)

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
