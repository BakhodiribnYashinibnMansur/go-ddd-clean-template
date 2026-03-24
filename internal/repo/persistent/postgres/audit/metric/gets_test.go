package metric

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
	t.Parallel()

	ctx := t.Context()
	now := time.Now()
	id1 := uuid.New()
	id2 := uuid.New()

	name := "UserService.Create"
	isPanicTrue := true
	isPanicFalse := false
	fromDate := now.Add(-24 * time.Hour)
	toDate := now
	panicErr := "stack overflow"

	tests := []struct {
		name          string
		filter        *domain.FunctionMetricsFilter
		setupMock     func(pgxmock.PgxPoolIface)
		expectedLen   int
		expectedCount int
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success with items",
			filter: &domain.FunctionMetricsFilter{
				Pagination: &domain.Pagination{Limit: 10, Offset: 0},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{
					"id", "name", "latency_ms", "is_panic", "panic_error", "created_at",
				}).
					AddRow(id1, "UserService.Create", 250, false, nil, now).
					AddRow(id2, "OrderService.Process", 10, true, &panicErr, now)
				mock.ExpectQuery("SELECT (.+) FROM function_metrics").
					WillReturnRows(rows)
			},
			expectedLen:   2,
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "success with all filters",
			filter: &domain.FunctionMetricsFilter{
				Name:       &name,
				IsPanic:    &isPanicFalse,
				FromDate:   &fromDate,
				ToDate:     &toDate,
				Pagination: &domain.Pagination{Limit: 5, Offset: 0},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(
						pgxmock.AnyArg(), // name
						pgxmock.AnyArg(), // is_panic
						pgxmock.AnyArg(), // from_date
						pgxmock.AnyArg(), // to_date
					).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{
					"id", "name", "latency_ms", "is_panic", "panic_error", "created_at",
				}).
					AddRow(id1, "UserService.Create", 250, false, nil, now)
				mock.ExpectQuery("SELECT (.+) FROM function_metrics").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnRows(rows)
			},
			expectedLen:   1,
			expectedCount: 1,
			expectedError: false,
		},
		{
			name: "success filter by is_panic true",
			filter: &domain.FunctionMetricsFilter{
				IsPanic:    &isPanicTrue,
				Pagination: &domain.Pagination{Limit: 10, Offset: 0},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{
					"id", "name", "latency_ms", "is_panic", "panic_error", "created_at",
				}).
					AddRow(id2, "OrderService.Process", 10, true, &panicErr, now)
				mock.ExpectQuery("SELECT (.+) FROM function_metrics").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedLen:   1,
			expectedCount: 1,
			expectedError: false,
		},
		{
			name: "empty result",
			filter: &domain.FunctionMetricsFilter{
				Pagination: &domain.Pagination{Limit: 10, Offset: 0},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{
					"id", "name", "latency_ms", "is_panic", "panic_error", "created_at",
				})
				mock.ExpectQuery("SELECT (.+) FROM function_metrics").
					WillReturnRows(rows)
			},
			expectedLen:   0,
			expectedCount: 0,
			expectedError: false,
		},
		{
			name: "count query error",
			filter: &domain.FunctionMetricsFilter{
				Pagination: &domain.Pagination{Limit: 10},
			},
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
			name: "select query error",
			filter: &domain.FunctionMetricsFilter{
				Pagination: &domain.Pagination{Limit: 10},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(countRows)

				mock.ExpectQuery("SELECT (.+) FROM function_metrics").
					WillReturnError(errors.New("database error"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "success with pagination offset",
			filter: &domain.FunctionMetricsFilter{
				Pagination: &domain.Pagination{Limit: 5, Offset: 10},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(25)
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{
					"id", "name", "latency_ms", "is_panic", "panic_error", "created_at",
				}).
					AddRow(id1, "CacheService.Flush", 5, false, nil, now)
				mock.ExpectQuery("SELECT (.+) FROM function_metrics").
					WillReturnRows(rows)
			},
			expectedLen:   1,
			expectedCount: 25,
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
				log:     logger.New("debug"),
			}

			metrics, count, err := repo.Gets(ctx, tt.filter)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, metrics, tt.expectedLen)
				assert.Equal(t, tt.expectedCount, count)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
