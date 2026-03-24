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

func TestRepo_Create(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	now := time.Now()

	rl := &domain.RateLimit{
		ID:            uuid.New(),
		Name:          "API Rate Limit",
		PathPattern:   "/api/v1/*",
		Method:        "GET",
		LimitCount:    100,
		WindowSeconds: 60,
		IsActive:      true,
	}

	tests := []struct {
		name          string
		rateLimit     *domain.RateLimit
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name:      "success",
			rateLimit: rl,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"created_at", "updated_at"}).
					AddRow(now, now)
				mock.ExpectQuery("INSERT INTO rate_limits").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // name
						pgxmock.AnyArg(), // path_pattern
						pgxmock.AnyArg(), // method
						pgxmock.AnyArg(), // limit_count
						pgxmock.AnyArg(), // window_seconds
						pgxmock.AnyArg(), // is_active
					).
					WillReturnRows(rows)
			},
			expectedError: false,
		},
		{
			name:      "database error",
			rateLimit: rl,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("INSERT INTO rate_limits").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // name
						pgxmock.AnyArg(), // path_pattern
						pgxmock.AnyArg(), // method
						pgxmock.AnyArg(), // limit_count
						pgxmock.AnyArg(), // window_seconds
						pgxmock.AnyArg(), // is_active
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
				logger:  logger.New("debug"),
			}

			err = repo.Create(ctx, tt.rateLimit)

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
