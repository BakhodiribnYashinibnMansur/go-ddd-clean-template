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

func TestRepo_Create(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	now := time.Now()
	panicErr := "runtime: goroutine stack overflow"

	tests := []struct {
		name          string
		metric        *domain.FunctionMetric
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success",
			metric: &domain.FunctionMetric{
				ID:         uuid.New(),
				Name:       "UserService.Create",
				LatencyMs:  250,
				IsPanic:    false,
				PanicError: nil,
				CreatedAt:  now,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO function_metrics").
					WithArgs(
						pgxmock.AnyArg(), // name
						pgxmock.AnyArg(), // latency_ms
						pgxmock.AnyArg(), // is_panic
						pgxmock.AnyArg(), // panic_error
						pgxmock.AnyArg(), // created_at
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: false,
		},
		{
			name: "success with panic",
			metric: &domain.FunctionMetric{
				ID:         uuid.New(),
				Name:       "OrderService.Process",
				LatencyMs:  10,
				IsPanic:    true,
				PanicError: &panicErr,
				CreatedAt:  now,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO function_metrics").
					WithArgs(
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
		{
			name: "database error",
			metric: &domain.FunctionMetric{
				ID:        uuid.New(),
				Name:      "AuthService.Login",
				LatencyMs: 100,
				CreatedAt: now,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO function_metrics").
					WithArgs(
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

			err = repo.Create(ctx, tt.metric)

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
