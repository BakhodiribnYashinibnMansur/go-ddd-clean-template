package scope

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Get(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	now := time.Now()

	tests := []struct {
		name          string
		filter        *domain.ScopeFilter
		setupMock     func(pgxmock.PgxPoolIface)
		expectedScope *domain.Scope
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success - get by path",
			filter: &domain.ScopeFilter{
				Path: func() *string { s := "/api/v1/users"; return &s }(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"path", "method", "created_at"}).
					AddRow("/api/v1/users", "GET", now)
				mock.ExpectQuery("SELECT (.+) FROM scope").
					WithArgs("/api/v1/users").
					WillReturnRows(rows)
			},
			expectedScope: &domain.Scope{
				Path:      "/api/v1/users",
				Method:    "GET",
				CreatedAt: now,
			},
			expectedError: false,
		},
		{
			name: "success - get by method",
			filter: &domain.ScopeFilter{
				Method: func() *string { s := "POST"; return &s }(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"path", "method", "created_at"}).
					AddRow("/api/v1/users", "POST", now)
				mock.ExpectQuery("SELECT (.+) FROM scope").
					WithArgs("POST").
					WillReturnRows(rows)
			},
			expectedScope: &domain.Scope{
				Path:      "/api/v1/users",
				Method:    "POST",
				CreatedAt: now,
			},
			expectedError: false,
		},
		{
			name: "success - get by path and method",
			filter: &domain.ScopeFilter{
				Path:   func() *string { s := "/api/v1/users"; return &s }(),
				Method: func() *string { s := "GET"; return &s }(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"path", "method", "created_at"}).
					AddRow("/api/v1/users", "GET", now)
				mock.ExpectQuery("SELECT (.+) FROM scope").
					WithArgs("/api/v1/users", "GET").
					WillReturnRows(rows)
			},
			expectedScope: &domain.Scope{
				Path:      "/api/v1/users",
				Method:    "GET",
				CreatedAt: now,
			},
			expectedError: false,
		},
		{
			name:   "empty filter",
			filter: &domain.ScopeFilter{},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"path", "method", "created_at"}).
					AddRow("/api/v1/users", "GET", now)
				mock.ExpectQuery("SELECT (.+) FROM scope").
					WillReturnRows(rows)
			},
			expectedScope: &domain.Scope{
				Path:      "/api/v1/users",
				Method:    "GET",
				CreatedAt: now,
			},
			expectedError: false,
		},
		{
			name: "not found",
			filter: &domain.ScopeFilter{
				Path: func() *string { s := "/api/v1/nonexistent"; return &s }(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM scope").
					WithArgs("/api/v1/nonexistent").
					WillReturnError(pgx.ErrNoRows)
			},
			expectedScope: nil,
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "database error",
			filter: &domain.ScopeFilter{
				Path: func() *string { s := "/api/v1/users"; return &s }(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM scope").
					WithArgs("/api/v1/users").
					WillReturnError(errors.New("database error"))
			},
			expectedScope: nil,
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "connection timeout",
			filter: &domain.ScopeFilter{
				Path: func() *string { s := "/api/v1/users"; return &s }(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM scope").
					WithArgs("/api/v1/users").
					WillReturnError(errors.New("connection timeout"))
			},
			expectedScope: nil,
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "timeout")
			},
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
				assert.Equal(t, tt.expectedScope.Path, result.Path)
				assert.Equal(t, tt.expectedScope.Method, result.Method)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
