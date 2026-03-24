package scope

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Create(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	now := time.Now()

	tests := []struct {
		name          string
		scope         *domain.Scope
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success",
			scope: &domain.Scope{
				Path:   "/api/v1/users",
				Method: "GET",
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"created_at"}).AddRow(now)
				mock.ExpectQuery("INSERT INTO scope").
					WithArgs(
						pgxmock.AnyArg(), // path
						pgxmock.AnyArg(), // method
						pgxmock.AnyArg(), // created_at
					).
					WillReturnRows(rows)
			},
			expectedError: false,
		},
		{
			name: "success - POST method",
			scope: &domain.Scope{
				Path:   "/api/v1/users",
				Method: "POST",
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"created_at"}).AddRow(now)
				mock.ExpectQuery("INSERT INTO scope").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnRows(rows)
			},
			expectedError: false,
		},
		{
			name: "success - DELETE method",
			scope: &domain.Scope{
				Path:   "/api/v1/users/:id",
				Method: "DELETE",
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"created_at"}).AddRow(now)
				mock.ExpectQuery("INSERT INTO scope").
					WithArgs(
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
			scope: &domain.Scope{
				Path:   "/api/v1/users",
				Method: "GET",
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("INSERT INTO scope").
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
			name: "duplicate scope error",
			scope: &domain.Scope{
				Path:   "/api/v1/users",
				Method: "GET",
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("INSERT INTO scope").
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
			name: "connection timeout",
			scope: &domain.Scope{
				Path:   "/api/v1/admin",
				Method: "PUT",
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("INSERT INTO scope").
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
		{
			name: "success - nested path",
			scope: &domain.Scope{
				Path:   "/api/v1/organizations/:orgId/teams/:teamId/members",
				Method: "GET",
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"created_at"}).AddRow(now)
				mock.ExpectQuery("INSERT INTO scope").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnRows(rows)
			},
			expectedError: false,
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

			err = repo.Create(ctx, tt.scope)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.False(t, tt.scope.CreatedAt.IsZero())
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
