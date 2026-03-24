package permission

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

	permID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	parentID := uuid.MustParse("770e8400-e29b-41d4-a716-446655440000")
	permName := "read_users"
	desc := "read users permission"

	tests := []struct {
		name          string
		filter        *domain.PermissionFilter
		setupMock     func(pgxmock.PgxPoolIface)
		expectedPerm  *domain.Permission
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success - get by id",
			filter: &domain.PermissionFilter{
				ID: &permID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "parent_id", "name", "description", "created_at"}).
					AddRow(permID, &parentID, permName, &desc, now)

				mock.ExpectQuery("SELECT (.+) FROM permission").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedPerm: &domain.Permission{
				ID:          permID,
				ParentID:    &parentID,
				Name:        permName,
				Description: &desc,
				CreatedAt:   now,
			},
			expectedError: false,
		},
		{
			name: "success - get by name",
			filter: &domain.PermissionFilter{
				Name: &permName,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "parent_id", "name", "description", "created_at"}).
					AddRow(permID, &parentID, permName, &desc, now)

				mock.ExpectQuery("SELECT (.+) FROM permission").
					WithArgs(permName).
					WillReturnRows(rows)
			},
			expectedPerm: &domain.Permission{
				ID:          permID,
				ParentID:    &parentID,
				Name:        permName,
				Description: &desc,
				CreatedAt:   now,
			},
			expectedError: false,
		},
		{
			name: "not found",
			filter: &domain.PermissionFilter{
				ID: &permID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM permission").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedPerm:  nil,
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "database error",
			filter: &domain.PermissionFilter{
				ID: &permID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM permission").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			expectedPerm:  nil,
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name:   "empty filter",
			filter: &domain.PermissionFilter{},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "parent_id", "name", "description", "created_at"}).
					AddRow(permID, &parentID, permName, &desc, now)

				mock.ExpectQuery("SELECT (.+) FROM permission").
					WillReturnRows(rows)
			},
			expectedPerm: &domain.Permission{
				ID:          permID,
				ParentID:    &parentID,
				Name:        permName,
				Description: &desc,
				CreatedAt:   now,
			},
			expectedError: false,
		},
		{
			name: "success - nil parent_id",
			filter: &domain.PermissionFilter{
				ID: &permID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "parent_id", "name", "description", "created_at"}).
					AddRow(permID, nil, permName, &desc, now)

				mock.ExpectQuery("SELECT (.+) FROM permission").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedPerm: &domain.Permission{
				ID:          permID,
				ParentID:    nil,
				Name:        permName,
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
				assert.Equal(t, tt.expectedPerm.ID, result.ID)
				assert.Equal(t, tt.expectedPerm.ParentID, result.ParentID)
				assert.Equal(t, tt.expectedPerm.Name, result.Name)
				assert.Equal(t, tt.expectedPerm.Description, result.Description)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
