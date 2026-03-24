package permission

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

	parentID := uuid.MustParse("770e8400-e29b-41d4-a716-446655440000")
	desc := "test permission"
	perm := &domain.Permission{
		ParentID:    &parentID,
		Name:        "read_users",
		Description: &desc,
	}

	tests := []struct {
		name          string
		perm          *domain.Permission
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success",
			perm: perm,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				returnedID := uuid.New()
				returnedAt := time.Now()
				rows := pgxmock.NewRows([]string{"id", "created_at"}).
					AddRow(returnedID, returnedAt)

				mock.ExpectQuery("INSERT INTO permission").
					WithArgs(
						pgxmock.AnyArg(), // parent_id
						pgxmock.AnyArg(), // name
						pgxmock.AnyArg(), // description
						pgxmock.AnyArg(), // created_at
					).
					WillReturnRows(rows)
			},
			expectedError: false,
		},
		{
			name: "success - nil parent_id",
			perm: &domain.Permission{
				ParentID:    nil,
				Name:        "root_perm",
				Description: &desc,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				returnedID := uuid.New()
				returnedAt := time.Now()
				rows := pgxmock.NewRows([]string{"id", "created_at"}).
					AddRow(returnedID, returnedAt)

				mock.ExpectQuery("INSERT INTO permission").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnRows(rows)
			},
			expectedError: false,
		},
		{
			name: "database error",
			perm: perm,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("INSERT INTO permission").
					WithArgs(
						pgxmock.AnyArg(),
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
			perm: perm,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("INSERT INTO permission").
					WithArgs(
						pgxmock.AnyArg(),
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

			err = repo.Create(ctx, tt.perm)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, tt.perm.ID)
				assert.False(t, tt.perm.CreatedAt.IsZero())
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
