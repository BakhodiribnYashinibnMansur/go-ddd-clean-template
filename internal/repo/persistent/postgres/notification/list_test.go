package notification

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

	tests := []struct {
		name          string
		filter        domain.NotificationFilter
		setupMock     func(pgxmock.PgxPoolIface)
		expectedCount int
		expectedTotal int64
		expectedError bool
	}{
		{
			name: "success with results",
			filter: domain.NotificationFilter{
				Limit:  10,
				Offset: 0,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(2)))

				now := time.Now()
				id1, id2 := uuid.New(), uuid.New()
				mock.ExpectQuery("SELECT (.+) FROM notifications").
					WillReturnRows(
						pgxmock.NewRows([]string{
							"id", "title", "body", "type", "target_type",
							"is_active", "created_at", "updated_at",
						}).
							AddRow(id1, "First", "Body 1", "info", "all", true, now, now).
							AddRow(id2, "Second", "Body 2", "warning", "user", false, now, now),
					)
			},
			expectedCount: 2,
			expectedTotal: 2,
			expectedError: false,
		},
		{
			name: "success with search filter",
			filter: domain.NotificationFilter{
				Search: "welcome",
				Limit:  10,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(1)))

				now := time.Now()
				mock.ExpectQuery("SELECT (.+) FROM notifications").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(
						pgxmock.NewRows([]string{
							"id", "title", "body", "type", "target_type",
							"is_active", "created_at", "updated_at",
						}).AddRow(uuid.New(), "Welcome", "Hello", "info", "all", true, now, now),
					)
			},
			expectedCount: 1,
			expectedTotal: 1,
			expectedError: false,
		},
		{
			name: "empty results",
			filter: domain.NotificationFilter{
				Search: "nonexistent",
				Limit:  10,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT COUNT").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(int64(0)))

				mock.ExpectQuery("SELECT (.+) FROM notifications").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(
						pgxmock.NewRows([]string{
							"id", "title", "body", "type", "target_type",
							"is_active", "created_at", "updated_at",
						}),
					)
			},
			expectedCount: 0,
			expectedTotal: 0,
			expectedError: false,
		},
		{
			name:   "count query error",
			filter: domain.NotificationFilter{Limit: 10},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT COUNT").
					WillReturnError(errors.New("database error"))
			},
			expectedCount: 0,
			expectedTotal: 0,
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

			items, total, err := repo.List(t.Context(), tt.filter)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedTotal, total)
				assert.Len(t, items, tt.expectedCount)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
