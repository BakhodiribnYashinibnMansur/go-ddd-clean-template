package scope

import (
	"errors"
	"testing"

	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Delete(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	tests := []struct {
		name          string
		path          string
		method        string
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name:   "success",
			path:   "/api/v1/users",
			method: "GET",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM scope").
					WithArgs("/api/v1/users", "GET").
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			expectedError: false,
		},
		{
			name:   "success - POST method",
			path:   "/api/v1/users",
			method: "POST",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM scope").
					WithArgs("/api/v1/users", "POST").
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			expectedError: false,
		},
		{
			name:   "not found - no rows affected",
			path:   "/api/v1/nonexistent",
			method: "GET",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM scope").
					WithArgs("/api/v1/nonexistent", "GET").
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			},
		},
		{
			name:   "database error",
			path:   "/api/v1/users",
			method: "GET",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM scope").
					WithArgs("/api/v1/users", "GET").
					WillReturnError(errors.New("database error"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name:   "foreign key constraint violation",
			path:   "/api/v1/users",
			method: "GET",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM scope").
					WithArgs("/api/v1/users", "GET").
					WillReturnError(errors.New("foreign key constraint violation"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "foreign key")
			},
		},
		{
			name:   "connection timeout",
			path:   "/api/v1/users",
			method: "GET",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM scope").
					WithArgs("/api/v1/users", "GET").
					WillReturnError(errors.New("connection timeout"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "timeout")
			},
		},
		{
			name:   "empty path and method",
			path:   "",
			method: "",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM scope").
					WithArgs("", "").
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			},
		},
		{
			name:   "success - nested path with DELETE method",
			path:   "/api/v1/organizations/:orgId/teams/:teamId",
			method: "DELETE",
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM scope").
					WithArgs("/api/v1/organizations/:orgId/teams/:teamId", "DELETE").
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
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

			err = repo.Delete(ctx, tt.path, tt.method)

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
