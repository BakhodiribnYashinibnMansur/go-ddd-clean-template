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

func TestRepo_Create(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	now := time.Now()

	userID := uuid.New()
	requestID := uuid.New()
	stackTrace := "goroutine 1 [running]:\nmain.main()\n\t/app/main.go:10"
	serviceName := "auth-service"
	ipAddr := "192.168.1.100"
	path := "/api/v1/users"
	method := "POST"

	sysErr := &domain.SystemError{
		ID:          uuid.New(),
		Code:        "ERR_AUTH_001",
		Message:     "authentication failed",
		StackTrace:  &stackTrace,
		Metadata:    map[string]any{"attempt": 3},
		Severity:    "ERROR",
		ServiceName: &serviceName,
		RequestID:   &requestID,
		UserID:      &userID,
		IPAddress:   &ipAddr,
		Path:        &path,
		Method:      &method,
		CreatedAt:   now,
	}

	tests := []struct {
		name          string
		sysErr        *domain.SystemError
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name:   "success",
			sysErr: sysErr,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO system_errors").
					WithArgs(
						pgxmock.AnyArg(), // code
						pgxmock.AnyArg(), // message
						pgxmock.AnyArg(), // stack_trace
						pgxmock.AnyArg(), // metadata
						pgxmock.AnyArg(), // severity
						pgxmock.AnyArg(), // service_name
						pgxmock.AnyArg(), // request_id
						pgxmock.AnyArg(), // user_id
						pgxmock.AnyArg(), // ip_address
						pgxmock.AnyArg(), // path
						pgxmock.AnyArg(), // method
						pgxmock.AnyArg(), // created_at
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: false,
		},
		{
			name:   "database error",
			sysErr: sysErr,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO system_errors").
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
			sysErr: &domain.SystemError{
				ID:        uuid.New(),
				Code:      "ERR_INTERNAL",
				Message:   "internal server error",
				Severity:  "FATAL",
				CreatedAt: now,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO system_errors").
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

			err = repo.Create(ctx, tt.sysErr)

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
