package iprule

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

	tests := []struct {
		name          string
		rule          *domain.IPRule
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
	}{
		{
			name: "success",
			rule: &domain.IPRule{
				ID:        uuid.New(),
				IPAddress: "192.168.1.100",
				Type:      "whitelist",
				Reason:    "Trusted office IP",
				IsActive:  true,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				now := time.Now()
				mock.ExpectQuery("INSERT INTO ip_rules").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // ip_address
						pgxmock.AnyArg(), // type
						pgxmock.AnyArg(), // reason
						pgxmock.AnyArg(), // is_active
					).
					WillReturnRows(
						pgxmock.NewRows([]string{"created_at", "updated_at"}).
							AddRow(now, now),
					)
			},
			expectedError: false,
		},
		{
			name: "database error",
			rule: &domain.IPRule{
				ID:        uuid.New(),
				IPAddress: "10.0.0.1",
				Type:      "blacklist",
				Reason:    "Suspicious",
				IsActive:  true,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("INSERT INTO ip_rules").
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

			err = repo.Create(t.Context(), tt.rule)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.False(t, tt.rule.CreatedAt.IsZero())
				assert.False(t, tt.rule.UpdatedAt.IsZero())
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
