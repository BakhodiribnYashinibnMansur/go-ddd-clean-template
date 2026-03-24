package role

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Create(t *testing.T) {
	ctx := t.Context()

	desc := "test role description"
	role := &domain.Role{
		Name:        "admin",
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
				returnedID := uuid.New()
				returnedAt := time.Now()
				rows := pgxmock.NewRows([]string{"id", "created_at"}).
					AddRow(returnedID, returnedAt)

				mock.ExpectQuery("INSERT INTO role").
					WithArgs(
						pgxmock.AnyArg(), // name
						pgxmock.AnyArg(), // description
						pgxmock.AnyArg(), // created_at
					).
					WillReturnRows(rows)
			},
			expectedError: false,
		},
		{
			name: "database error",
			role: role,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("INSERT INTO role").
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
			name: "duplicate name error",
			role: role,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("INSERT INTO role").
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

			err = repo.Create(ctx, tt.role)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, tt.role.ID)
				assert.False(t, tt.role.CreatedAt.IsZero())
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
