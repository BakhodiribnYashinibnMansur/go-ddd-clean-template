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

func TestRepo_Gets(t *testing.T) {
	ctx := t.Context()
	now := time.Now()

	permID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	parentID := uuid.MustParse("770e8400-e29b-41d4-a716-446655440000")
	permName := "read_users"
	desc := "read users permission"

	tests := []struct {
		name          string
		filter        *domain.PermissionsFilter
		setupMock     func(pgxmock.PgxPoolIface)
		expectedPerms []*domain.Permission
		expectedCount int
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success - get all permissions",
			filter: &domain.PermissionsFilter{
				PermissionFilter: domain.PermissionFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				desc2 := "write users permission"
				rows := pgxmock.NewRows([]string{"id", "parent_id", "name", "description", "created_at"}).
					AddRow(permID, &parentID, permName, &desc, now).
					AddRow(uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), nil, "write_users", &desc2, now)

				mock.ExpectQuery("SELECT (.+) FROM permission").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(2))
			},
			expectedPerms: []*domain.Permission{
				{ID: permID, ParentID: &parentID, Name: permName, Description: &desc, CreatedAt: now},
				{ID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), Name: "write_users"},
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "success - empty result",
			filter: &domain.PermissionsFilter{
				PermissionFilter: domain.PermissionFilter{
					Name: func() *string { s := "nonexistent"; return &s }(),
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "parent_id", "name", "description", "created_at"})

				mock.ExpectQuery("SELECT (.+) FROM permission").
					WithArgs("nonexistent").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WithArgs("nonexistent").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedPerms: nil,
			expectedCount: 0,
			expectedError: false,
		},
		{
			name: "success - with pagination",
			filter: &domain.PermissionsFilter{
				PermissionFilter: domain.PermissionFilter{},
				Pagination: &domain.Pagination{
					Limit:  10,
					Offset: 0,
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "parent_id", "name", "description", "created_at"}).
					AddRow(permID, &parentID, permName, &desc, now)

				mock.ExpectQuery("SELECT (.+) FROM permission").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(50))
			},
			expectedPerms: []*domain.Permission{
				{ID: permID, ParentID: &parentID, Name: permName, Description: &desc, CreatedAt: now},
			},
			expectedCount: 50,
			expectedError: false,
		},
		{
			name: "database error on select",
			filter: &domain.PermissionsFilter{
				PermissionFilter: domain.PermissionFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM permission").
					WillReturnError(errors.New("database error"))
			},
			expectedPerms: nil,
			expectedCount: 0,
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "database error on count",
			filter: &domain.PermissionsFilter{
				PermissionFilter: domain.PermissionFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "parent_id", "name", "description", "created_at"})
				mock.ExpectQuery("SELECT (.+) FROM permission").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WillReturnError(errors.New("count query error"))
			},
			expectedPerms: nil,
			expectedCount: 0,
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "success - filter by id",
			filter: &domain.PermissionsFilter{
				PermissionFilter: domain.PermissionFilter{
					ID: &permID,
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "parent_id", "name", "description", "created_at"}).
					AddRow(permID, &parentID, permName, &desc, now)

				mock.ExpectQuery("SELECT (.+) FROM permission").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
			},
			expectedPerms: []*domain.Permission{
				{ID: permID, ParentID: &parentID, Name: permName, Description: &desc, CreatedAt: now},
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

			perms, count, err := repo.Gets(ctx, tt.filter)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, perms)
				assert.Equal(t, 0, count)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
				assert.Equal(t, len(tt.expectedPerms), len(perms))
				for i, expected := range tt.expectedPerms {
					assert.Equal(t, expected.ID, perms[i].ID)
					assert.Equal(t, expected.Name, perms[i].Name)
				}
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
