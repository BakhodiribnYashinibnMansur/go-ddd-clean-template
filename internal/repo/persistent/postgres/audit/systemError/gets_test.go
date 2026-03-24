package systemerror

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
	userID := uuid.New()
	requestID := uuid.New()
	resolvedBy := uuid.New()
	resolvedAt := now.Add(-1 * time.Hour)

	stackTrace := "goroutine 1 [running]"
	serviceName := "auth-service"
	ipAddr := "10.0.0.1"
	path := "/api/v1/users"
	method := "POST"
	code := "ERR_AUTH_001"
	severity := "ERROR"
	isResolvedTrue := true
	isResolvedFalse := false
	fromDate := now.Add(-24 * time.Hour)
	toDate := now

	tests := []struct {
		name          string
		filter        *domain.SystemErrorsFilter
		setupMock     func(pgxmock.PgxPoolIface)
		expectedLen   int
		expectedCount int
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success with items",
			filter: &domain.SystemErrorsFilter{
				Pagination: &domain.Pagination{Limit: 10, Offset: 0},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{
					"id", "code", "message", "stack_trace", "metadata",
					"severity", "service_name", "request_id", "user_id",
					"ip_address", "path", "method",
					"is_resolved", "resolved_at", "resolved_by", "created_at",
				}).
					AddRow(id1, "ERR_AUTH_001", "auth failed", &stackTrace,
						map[string]any{"attempt": 3}, "ERROR", &serviceName, &requestID,
						&userID, &ipAddr, &path, &method,
						true, &resolvedAt, &resolvedBy, now).
					AddRow(id2, "ERR_INTERNAL", "server error", nil,
						nil, "FATAL", nil, nil,
						nil, nil, nil, nil,
						false, nil, nil, now)
				mock.ExpectQuery("SELECT (.+) FROM system_errors").
					WillReturnRows(rows)
			},
			expectedLen:   2,
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "success with all filters",
			filter: &domain.SystemErrorsFilter{
				SystemErrorFilter: domain.SystemErrorFilter{
					Code:       &code,
					Severity:   &severity,
					IsResolved: &isResolvedFalse,
					RequestID:  &requestID,
					UserID:     &userID,
					FromDate:   &fromDate,
					ToDate:     &toDate,
				},
				Pagination: &domain.Pagination{Limit: 5, Offset: 0},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(
						pgxmock.AnyArg(), // code
						pgxmock.AnyArg(), // severity
						pgxmock.AnyArg(), // is_resolved
						pgxmock.AnyArg(), // request_id
						pgxmock.AnyArg(), // user_id
						pgxmock.AnyArg(), // from_date
						pgxmock.AnyArg(), // to_date
					).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{
					"id", "code", "message", "stack_trace", "metadata",
					"severity", "service_name", "request_id", "user_id",
					"ip_address", "path", "method",
					"is_resolved", "resolved_at", "resolved_by", "created_at",
				}).
					AddRow(id1, "ERR_AUTH_001", "auth failed", &stackTrace,
						map[string]any{"attempt": 3}, "ERROR", &serviceName, &requestID,
						&userID, &ipAddr, &path, &method,
						false, nil, nil, now)
				mock.ExpectQuery("SELECT (.+) FROM system_errors").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
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
			name: "success filter by is_resolved true",
			filter: &domain.SystemErrorsFilter{
				SystemErrorFilter: domain.SystemErrorFilter{
					IsResolved: &isResolvedTrue,
				},
				Pagination: &domain.Pagination{Limit: 10, Offset: 0},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{
					"id", "code", "message", "stack_trace", "metadata",
					"severity", "service_name", "request_id", "user_id",
					"ip_address", "path", "method",
					"is_resolved", "resolved_at", "resolved_by", "created_at",
				}).
					AddRow(id1, "ERR_AUTH_001", "auth failed", nil,
						nil, "ERROR", nil, nil,
						nil, nil, nil, nil,
						true, &resolvedAt, &resolvedBy, now)
				mock.ExpectQuery("SELECT (.+) FROM system_errors").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedLen:   1,
			expectedCount: 1,
			expectedError: false,
		},
		{
			name: "empty result",
			filter: &domain.SystemErrorsFilter{
				Pagination: &domain.Pagination{Limit: 10, Offset: 0},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{
					"id", "code", "message", "stack_trace", "metadata",
					"severity", "service_name", "request_id", "user_id",
					"ip_address", "path", "method",
					"is_resolved", "resolved_at", "resolved_by", "created_at",
				})
				mock.ExpectQuery("SELECT (.+) FROM system_errors").
					WillReturnRows(rows)
			},
			expectedLen:   0,
			expectedCount: 0,
			expectedError: false,
		},
		{
			name: "count query error",
			filter: &domain.SystemErrorsFilter{
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
			filter: &domain.SystemErrorsFilter{
				Pagination: &domain.Pagination{Limit: 10},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(countRows)

				mock.ExpectQuery("SELECT (.+) FROM system_errors").
					WillReturnError(errors.New("database error"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "success with pagination offset",
			filter: &domain.SystemErrorsFilter{
				Pagination: &domain.Pagination{Limit: 5, Offset: 10},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(30)
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{
					"id", "code", "message", "stack_trace", "metadata",
					"severity", "service_name", "request_id", "user_id",
					"ip_address", "path", "method",
					"is_resolved", "resolved_at", "resolved_by", "created_at",
				}).
					AddRow(id1, "ERR_DB", "connection lost", nil,
						nil, "FATAL", nil, nil,
						nil, nil, nil, nil,
						false, nil, nil, now)
				mock.ExpectQuery("SELECT (.+) FROM system_errors").
					WillReturnRows(rows)
			},
			expectedLen:   1,
			expectedCount: 30,
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

			sysErrors, count, err := repo.Gets(ctx, tt.filter)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, sysErrors, tt.expectedLen)
				assert.Equal(t, tt.expectedCount, count)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
