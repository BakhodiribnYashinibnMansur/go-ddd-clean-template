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

func TestRepo_Create(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	now := time.Now()

	userID := uuid.New()
	sessionID := uuid.New()
	requestID := uuid.New()
	platform := "mobile"
	ipAddr := "10.0.0.1"
	userAgent := "TestAgent/1.0"
	permission := "orders:read"
	decision := "ALLOW"
	respSize := 1024
	errMsg := "not found"

	history := &domain.EndpointHistory{
		ID:           uuid.New(),
		UserID:       &userID,
		SessionID:    &sessionID,
		Method:       "GET",
		Path:         "/api/v1/users",
		StatusCode:   200,
		DurationMs:   150,
		Platform:     &platform,
		IPAddress:    &ipAddr,
		UserAgent:    &userAgent,
		Permission:   &permission,
		Decision:     &decision,
		RequestID:    &requestID,
		RateLimited:  false,
		ResponseSize: &respSize,
		ErrorMessage: &errMsg,
		CreatedAt:    now,
	}

	tests := []struct {
		name          string
		history       *domain.EndpointHistory
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name:    "success",
			history: history,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO endpoint_history").
					WithArgs(
						pgxmock.AnyArg(), // user_id
						pgxmock.AnyArg(), // session_id
						pgxmock.AnyArg(), // method
						pgxmock.AnyArg(), // path
						pgxmock.AnyArg(), // status_code
						pgxmock.AnyArg(), // duration_ms
						pgxmock.AnyArg(), // platform
						pgxmock.AnyArg(), // ip_address
						pgxmock.AnyArg(), // user_agent
						pgxmock.AnyArg(), // permission
						pgxmock.AnyArg(), // decision
						pgxmock.AnyArg(), // request_id
						pgxmock.AnyArg(), // rate_limited
						pgxmock.AnyArg(), // response_size
						pgxmock.AnyArg(), // error_message
						pgxmock.AnyArg(), // created_at
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: false,
		},
		{
			name:    "database error",
			history: history,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO endpoint_history").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
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
			name: "success with nil optional fields",
			history: &domain.EndpointHistory{
				ID:         uuid.New(),
				Method:     "POST",
				Path:       "/api/v1/login",
				StatusCode: 401,
				DurationMs: 50,
				CreatedAt:  now,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO endpoint_history").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
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
				log:     logger.New("debug"),
			}

			err = repo.Create(ctx, tt.history)

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
