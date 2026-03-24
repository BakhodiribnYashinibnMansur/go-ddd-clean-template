package history

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
	sessionID := uuid.New()
	requestID := uuid.New()

	platform := "web"
	ipAddr := "10.0.0.1"
	userAgent := "TestAgent"
	permission := "users:read"
	decision := "ALLOW"
	respSize := 2048
	method := "GET"
	path := "/api/v1/users"
	statusCode := 200
	fromDate := now.Add(-24 * time.Hour)
	toDate := now

	tests := []struct {
		name          string
		filter        *domain.EndpointHistoriesFilter
		setupMock     func(pgxmock.PgxPoolIface)
		expectedLen   int
		expectedCount int
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success with items",
			filter: &domain.EndpointHistoriesFilter{
				Pagination: &domain.Pagination{Limit: 10, Offset: 0},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{
					"id", "user_id", "session_id", "method", "path",
					"status_code", "duration_ms", "platform", "ip_address",
					"user_agent", "permission", "decision", "request_id",
					"rate_limited", "response_size", "error_message", "created_at",
				}).
					AddRow(id1, &userID, &sessionID, "GET", "/api/v1/users",
						200, 150, &platform, &ipAddr,
						&userAgent, &permission, &decision, &requestID,
						false, &respSize, nil, now).
					AddRow(id2, nil, nil, "POST", "/api/v1/login",
						401, 50, nil, nil,
						nil, nil, nil, nil,
						false, nil, nil, now)
				mock.ExpectQuery("SELECT (.+) FROM endpoint_history").
					WillReturnRows(rows)
			},
			expectedLen:   2,
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "success with all filters",
			filter: &domain.EndpointHistoriesFilter{
				EndpointHistoryFilter: domain.EndpointHistoryFilter{
					UserID:     &userID,
					Method:     &method,
					Path:       &path,
					StatusCode: &statusCode,
					FromDate:   &fromDate,
					ToDate:     &toDate,
				},
				Pagination: &domain.Pagination{Limit: 5, Offset: 0},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(
						pgxmock.AnyArg(), // user_id
						pgxmock.AnyArg(), // method
						pgxmock.AnyArg(), // path
						pgxmock.AnyArg(), // status_code
						pgxmock.AnyArg(), // from_date
						pgxmock.AnyArg(), // to_date
					).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{
					"id", "user_id", "session_id", "method", "path",
					"status_code", "duration_ms", "platform", "ip_address",
					"user_agent", "permission", "decision", "request_id",
					"rate_limited", "response_size", "error_message", "created_at",
				}).
					AddRow(id1, &userID, &sessionID, "GET", "/api/v1/users",
						200, 150, &platform, &ipAddr,
						&userAgent, &permission, &decision, &requestID,
						false, &respSize, nil, now)
				mock.ExpectQuery("SELECT (.+) FROM endpoint_history").
					WithArgs(
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
			name: "empty result",
			filter: &domain.EndpointHistoriesFilter{
				Pagination: &domain.Pagination{Limit: 10, Offset: 0},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{
					"id", "user_id", "session_id", "method", "path",
					"status_code", "duration_ms", "platform", "ip_address",
					"user_agent", "permission", "decision", "request_id",
					"rate_limited", "response_size", "error_message", "created_at",
				})
				mock.ExpectQuery("SELECT (.+) FROM endpoint_history").
					WillReturnRows(rows)
			},
			expectedLen:   0,
			expectedCount: 0,
			expectedError: false,
		},
		{
			name: "count query error",
			filter: &domain.EndpointHistoriesFilter{
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
			filter: &domain.EndpointHistoriesFilter{
				Pagination: &domain.Pagination{Limit: 10},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(countRows)

				mock.ExpectQuery("SELECT (.+) FROM endpoint_history").
					WillReturnError(errors.New("database error"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "success with pagination offset",
			filter: &domain.EndpointHistoriesFilter{
				Pagination: &domain.Pagination{Limit: 5, Offset: 10},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(20)
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{
					"id", "user_id", "session_id", "method", "path",
					"status_code", "duration_ms", "platform", "ip_address",
					"user_agent", "permission", "decision", "request_id",
					"rate_limited", "response_size", "error_message", "created_at",
				}).
					AddRow(id1, nil, nil, "DELETE", "/api/v1/users/1",
						204, 30, nil, nil,
						nil, nil, nil, nil,
						false, nil, nil, now)
				mock.ExpectQuery("SELECT (.+) FROM endpoint_history").
					WillReturnRows(rows)
			},
			expectedLen:   1,
			expectedCount: 20,
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

			histories, count, err := repo.Gets(ctx, tt.filter)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, histories, tt.expectedLen)
				assert.Equal(t, tt.expectedCount, count)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
