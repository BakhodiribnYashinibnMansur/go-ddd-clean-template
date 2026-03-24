package ratelimit

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

func TestRepo_List(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	now := time.Now()
	id1 := uuid.New()
	id2 := uuid.New()

	tests := []struct {
		name          string
		filter        domain.RateLimitFilter
		setupMock     func(pgxmock.PgxPoolIface)
		expectedLen   int
		expectedTotal int64
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name:   "success with items",
			filter: domain.RateLimitFilter{Limit: 10, Offset: 0},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(2))
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{
					"id", "name", "path_pattern", "method", "limit_count", "window_seconds",
					"is_active", "created_at", "updated_at",
				}).
					AddRow(id1, "Rate Limit 1", "/api/v1/*", "GET", 100, 60, true, now, now).
					AddRow(id2, "Rate Limit 2", "/api/v2/*", "POST", 50, 30, false, now, now)
				mock.ExpectQuery("SELECT (.+) FROM rate_limits").
					WillReturnRows(rows)
			},
			expectedLen:   2,
			expectedTotal: 2,
			expectedError: false,
		},
		{
			name:   "empty list",
			filter: domain.RateLimitFilter{Limit: 10, Offset: 0},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(0))
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{
					"id", "name", "path_pattern", "method", "limit_count", "window_seconds",
					"is_active", "created_at", "updated_at",
				})
				mock.ExpectQuery("SELECT (.+) FROM rate_limits").
					WillReturnRows(rows)
			},
			expectedLen:   0,
			expectedTotal: 0,
			expectedError: false,
		},
		{
			name:   "count query error",
			filter: domain.RateLimitFilter{Limit: 10},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT COUNT").
					WillReturnError(errors.New("database error"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name:   "list query error",
			filter: domain.RateLimitFilter{Limit: 10},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(countRows)

				mock.ExpectQuery("SELECT (.+) FROM rate_limits").
					WillReturnError(errors.New("database error"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
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

			items, total, err := repo.List(ctx, tt.filter)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, items, tt.expectedLen)
				assert.Equal(t, tt.expectedTotal, total)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
