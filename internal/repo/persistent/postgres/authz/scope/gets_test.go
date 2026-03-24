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

func TestRepo_Gets(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	now := time.Now()

	tests := []struct {
		name           string
		filter         *domain.ScopesFilter
		setupMock      func(pgxmock.PgxPoolIface)
		expectedScopes int
		expectedCount  int
		expectedError  bool
		errorCheck     func(*testing.T, error)
	}{
		{
			name: "success - get all",
			filter: &domain.ScopesFilter{
				ScopeFilter: domain.ScopeFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"path", "method", "created_at"}).
					AddRow("/api/v1/users", "GET", now).
					AddRow("/api/v1/users", "POST", now).
					AddRow("/api/v1/admin", "GET", now)

				mock.ExpectQuery("SELECT (.+) FROM scope").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(3))
			},
			expectedScopes: 3,
			expectedCount:  3,
			expectedError:  false,
		},
		{
			name: "success - filter by path",
			filter: &domain.ScopesFilter{
				ScopeFilter: domain.ScopeFilter{
					Path: func() *string { s := "/api/v1/users"; return &s }(),
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"path", "method", "created_at"}).
					AddRow("/api/v1/users", "GET", now).
					AddRow("/api/v1/users", "POST", now)

				mock.ExpectQuery("SELECT (.+) FROM scope").
					WithArgs("/api/v1/users").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WithArgs("/api/v1/users").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(2))
			},
			expectedScopes: 2,
			expectedCount:  2,
			expectedError:  false,
		},
		{
			name: "success - filter by method",
			filter: &domain.ScopesFilter{
				ScopeFilter: domain.ScopeFilter{
					Method: func() *string { s := "GET"; return &s }(),
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"path", "method", "created_at"}).
					AddRow("/api/v1/users", "GET", now)

				mock.ExpectQuery("SELECT (.+) FROM scope").
					WithArgs("GET").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WithArgs("GET").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
			},
			expectedScopes: 1,
			expectedCount:  1,
			expectedError:  false,
		},
		{
			name: "success - with pagination",
			filter: &domain.ScopesFilter{
				ScopeFilter: domain.ScopeFilter{},
				Pagination:  &domain.Pagination{Limit: 10, Offset: 0},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"path", "method", "created_at"}).
					AddRow("/api/v1/users", "GET", now)

				mock.ExpectQuery("SELECT (.+) FROM scope").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(50))
			},
			expectedScopes: 1,
			expectedCount:  50,
			expectedError:  false,
		},
		{
			name: "success - empty result",
			filter: &domain.ScopesFilter{
				ScopeFilter: domain.ScopeFilter{
					Path: func() *string { s := "/api/v1/nonexistent"; return &s }(),
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"path", "method", "created_at"})

				mock.ExpectQuery("SELECT (.+) FROM scope").
					WithArgs("/api/v1/nonexistent").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WithArgs("/api/v1/nonexistent").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedScopes: 0,
			expectedCount:  0,
			expectedError:  false,
		},
		{
			name: "database error on query",
			filter: &domain.ScopesFilter{
				ScopeFilter: domain.ScopeFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM scope").
					WillReturnError(errors.New("database error"))
			},
			expectedScopes: 0,
			expectedCount:  0,
			expectedError:  true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "database error on count",
			filter: &domain.ScopesFilter{
				ScopeFilter: domain.ScopeFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"path", "method", "created_at"}).
					AddRow("/api/v1/users", "GET", now)

				mock.ExpectQuery("SELECT (.+) FROM scope").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WillReturnError(errors.New("count error"))
			},
			expectedScopes: 0,
			expectedCount:  0,
			expectedError:  true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "connection timeout",
			filter: &domain.ScopesFilter{
				ScopeFilter: domain.ScopeFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM scope").
					WillReturnError(errors.New("connection timeout"))
			},
			expectedScopes: 0,
			expectedCount:  0,
			expectedError:  true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "timeout")
			},
		},
		{
			name: "success - filter by path and method",
			filter: &domain.ScopesFilter{
				ScopeFilter: domain.ScopeFilter{
					Path:   func() *string { s := "/api/v1/users"; return &s }(),
					Method: func() *string { s := "GET"; return &s }(),
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"path", "method", "created_at"}).
					AddRow("/api/v1/users", "GET", now)

				mock.ExpectQuery("SELECT (.+) FROM scope").
					WithArgs("/api/v1/users", "GET").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WithArgs("/api/v1/users", "GET").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
			},
			expectedScopes: 1,
			expectedCount:  1,
			expectedError:  false,
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

			scopes, count, err := repo.Gets(ctx, tt.filter)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, scopes)
				assert.Equal(t, 0, count)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
				assert.Len(t, scopes, tt.expectedScopes)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
