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

func TestRepo_Update(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	now := time.Now()
	updatedAt := now.Add(time.Second)

	rl := &domain.RateLimit{
		ID:            uuid.New(),
		Name:          "Updated Rate Limit",
		PathPattern:   "/api/v2/*",
		Method:        "POST",
		LimitCount:    200,
		WindowSeconds: 120,
		IsActive:      false,
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
				rows := pgxmock.NewRows([]string{"updated_at"}).AddRow(updatedAt)
				mock.ExpectQuery("UPDATE rate_limits").
					WithArgs(
						pgxmock.AnyArg(), // name
						pgxmock.AnyArg(), // path_pattern
						pgxmock.AnyArg(), // method
						pgxmock.AnyArg(), // limit_count
						pgxmock.AnyArg(), // window_seconds
						pgxmock.AnyArg(), // is_active
						pgxmock.AnyArg(), // id (WHERE)
					).
					WillReturnRows(rows)
			},
			expectedError: false,
		},
		{
			name:      "database error",
			rateLimit: rl,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("UPDATE rate_limits").
					WithArgs(
						pgxmock.AnyArg(), // name
						pgxmock.AnyArg(), // path_pattern
						pgxmock.AnyArg(), // method
						pgxmock.AnyArg(), // limit_count
						pgxmock.AnyArg(), // window_seconds
						pgxmock.AnyArg(), // is_active
						pgxmock.AnyArg(), // id (WHERE)
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

			err = repo.Update(ctx, tt.rateLimit)

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
