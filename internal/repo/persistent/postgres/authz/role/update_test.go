package role

import (
	"errors"
	"testing"

	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Update(t *testing.T) {
	ctx := t.Context()

	roleID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	desc := "updated description"
	role := &domain.Role{
		ID:          roleID,
		Name:        "updated_admin",
		Description: &desc,
	}

	tests := []struct {
		name          string
		role          *domain.Role
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success",
			role: role,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE role").
					WithArgs(
						pgxmock.AnyArg(), // name
						pgxmock.AnyArg(), // description
						pgxmock.AnyArg(), // id (WHERE)
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: false,
		},
		{
			name: "not found - no rows affected",
			role: role,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE role").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			},
		},
		{
			name: "database error",
			role: role,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE role").
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

			err = repo.Update(ctx, tt.role)

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
