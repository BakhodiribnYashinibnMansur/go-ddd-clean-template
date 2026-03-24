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

func TestRepo_GetPermissions(t *testing.T) {
	ctx := t.Context()
	now := time.Now()

	roleID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	permID1 := uuid.MustParse("660e8400-e29b-41d4-a716-446655440001")
	permID2 := uuid.MustParse("660e8400-e29b-41d4-a716-446655440002")

	tests := []struct {
		name                string
		roleID              uuid.UUID
		setupMock           func(pgxmock.PgxPoolIface)
		expectedPermissions []*domain.Permission
		expectedError       bool
		errorCheck          func(*testing.T, error)
	}{
		{
			name:   "success - multiple permissions",
			roleID: roleID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "created_at"}).
					AddRow(permID1, "read_users", now).
					AddRow(permID2, "write_users", now)

				mock.ExpectQuery("SELECT (.+) FROM permission").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedPermissions: []*domain.Permission{
				{ID: permID1, Name: "read_users", CreatedAt: now},
				{ID: permID2, Name: "write_users", CreatedAt: now},
			},
			expectedError: false,
		},
		{
			name:   "success - empty result",
			roleID: roleID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "created_at"})

				mock.ExpectQuery("SELECT (.+) FROM permission").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedPermissions: nil,
			expectedError:       false,
		},
		{
			name:   "database error",
			roleID: roleID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM permission").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			expectedPermissions: nil,
			expectedError:       true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name:   "success - single permission",
			roleID: roleID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "created_at"}).
					AddRow(permID1, "admin_access", now)

				mock.ExpectQuery("SELECT (.+) FROM permission").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedPermissions: []*domain.Permission{
				{ID: permID1, Name: "admin_access", CreatedAt: now},
			},
			expectedError: false,
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

			permissions, err := repo.GetPermissions(ctx, tt.roleID)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, permissions)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, len(tt.expectedPermissions), len(permissions))
				for i, expected := range tt.expectedPermissions {
					assert.Equal(t, expected.ID, permissions[i].ID)
					assert.Equal(t, expected.Name, permissions[i].Name)
				}
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
