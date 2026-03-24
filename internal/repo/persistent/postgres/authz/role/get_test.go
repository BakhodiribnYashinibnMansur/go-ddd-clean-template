package role

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Get(t *testing.T) {
	ctx := t.Context()
	now := time.Now()

	roleID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	roleName := "admin"
	desc := "admin role"

	tests := []struct {
		name          string
		filter        *domain.RoleFilter
		setupMock     func(pgxmock.PgxPoolIface)
		expectedRole  *domain.Role
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success - get by id",
			filter: &domain.RoleFilter{
				ID: &roleID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "created_at"}).
					AddRow(roleID, roleName, &desc, now)

				mock.ExpectQuery("SELECT (.+) FROM role").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedRole: &domain.Role{
				ID:          roleID,
				Name:        roleName,
				Description: &desc,
				CreatedAt:   now,
			},
			expectedError: false,
		},
		{
			name: "success - get by name",
			filter: &domain.RoleFilter{
				Name: &roleName,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "created_at"}).
					AddRow(roleID, roleName, &desc, now)

				mock.ExpectQuery("SELECT (.+) FROM role").
					WithArgs(roleName).
					WillReturnRows(rows)
			},
			expectedRole: &domain.Role{
				ID:          roleID,
				Name:        roleName,
				Description: &desc,
				CreatedAt:   now,
			},
			expectedError: false,
		},
		{
			name: "not found",
			filter: &domain.RoleFilter{
				ID: &roleID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM role").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedRole:  nil,
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "database error",
			filter: &domain.RoleFilter{
				ID: &roleID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM role").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			expectedRole:  nil,
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name:   "empty filter",
			filter: &domain.RoleFilter{},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "created_at"}).
					AddRow(roleID, roleName, &desc, now)

				mock.ExpectQuery("SELECT (.+) FROM role").
					WillReturnRows(rows)
			},
			expectedRole: &domain.Role{
				ID:          roleID,
				Name:        roleName,
				Description: &desc,
				CreatedAt:   now,
			},
			expectedError: false,
		},
		{
			name: "success - get by id and name",
			filter: &domain.RoleFilter{
				ID:   &roleID,
				Name: &roleName,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "created_at"}).
					AddRow(roleID, roleName, &desc, now)

				mock.ExpectQuery("SELECT (.+) FROM role").
					WithArgs(pgxmock.AnyArg(), roleName).
					WillReturnRows(rows)
			},
			expectedRole: &domain.Role{
				ID:          roleID,
				Name:        roleName,
				Description: &desc,
				CreatedAt:   now,
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

			result, err := repo.Get(ctx, tt.filter)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, result)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedRole.ID, result.ID)
				assert.Equal(t, tt.expectedRole.Name, result.Name)
				assert.Equal(t, tt.expectedRole.Description, result.Description)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
