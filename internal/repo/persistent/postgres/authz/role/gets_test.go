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

func TestRepo_Gets(t *testing.T) {
	ctx := t.Context()
	now := time.Now()

	roleID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	roleName := "admin"
	desc := "admin role"

	tests := []struct {
		name          string
		filter        *domain.RolesFilter
		setupMock     func(pgxmock.PgxPoolIface)
		expectedRoles []*domain.Role
		expectedCount int
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success - get all roles",
			filter: &domain.RolesFilter{
				RoleFilter: domain.RoleFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				desc2 := "editor role"
				rows := pgxmock.NewRows([]string{"id", "name", "description", "created_at"}).
					AddRow(roleID, roleName, &desc, now).
					AddRow(uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), "editor", &desc2, now)

				mock.ExpectQuery("SELECT (.+) FROM role").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(2))
			},
			expectedRoles: []*domain.Role{
				{ID: roleID, Name: roleName, Description: &desc, CreatedAt: now},
				{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), Name: "editor"},
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "success - empty result",
			filter: &domain.RolesFilter{
				RoleFilter: domain.RoleFilter{
					Name: func() *string { s := "nonexistent"; return &s }(),
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "created_at"})

				mock.ExpectQuery("SELECT (.+) FROM role").
					WithArgs("nonexistent").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WithArgs("nonexistent").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedRoles: nil,
			expectedCount: 0,
			expectedError: false,
		},
		{
			name: "success - with pagination",
			filter: &domain.RolesFilter{
				RoleFilter: domain.RoleFilter{},
				Pagination: &domain.Pagination{
					Limit:  10,
					Offset: 0,
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "created_at"}).
					AddRow(roleID, roleName, &desc, now)

				mock.ExpectQuery("SELECT (.+) FROM role").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(25))
			},
			expectedRoles: []*domain.Role{
				{ID: roleID, Name: roleName, Description: &desc, CreatedAt: now},
			},
			expectedCount: 25,
			expectedError: false,
		},
		{
			name: "database error on select",
			filter: &domain.RolesFilter{
				RoleFilter: domain.RoleFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM role").
					WillReturnError(errors.New("database error"))
			},
			expectedRoles: nil,
			expectedCount: 0,
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "database error on count",
			filter: &domain.RolesFilter{
				RoleFilter: domain.RoleFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "created_at"})
				mock.ExpectQuery("SELECT (.+) FROM role").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WillReturnError(errors.New("count query error"))
			},
			expectedRoles: nil,
			expectedCount: 0,
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "success - filter by id",
			filter: &domain.RolesFilter{
				RoleFilter: domain.RoleFilter{
					ID: &roleID,
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "name", "description", "created_at"}).
					AddRow(roleID, roleName, &desc, now)

				mock.ExpectQuery("SELECT (.+) FROM role").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
			},
			expectedRoles: []*domain.Role{
				{ID: roleID, Name: roleName, Description: &desc, CreatedAt: now},
			},
			expectedCount: 1,
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

			roles, count, err := repo.Gets(ctx, tt.filter)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, roles)
				assert.Equal(t, 0, count)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
				assert.Equal(t, len(tt.expectedRoles), len(roles))
				for i, expected := range tt.expectedRoles {
					assert.Equal(t, expected.ID, roles[i].ID)
					assert.Equal(t, expected.Name, roles[i].Name)
				}
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
