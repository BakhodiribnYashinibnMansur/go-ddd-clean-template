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

func TestRepo_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		notification  *domain.Notification
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
	}{
		{
			name: "success",
			notification: &domain.Notification{
				ID:         uuid.New(),
				Title:      "Updated Title",
				Body:       "Updated Body",
				Type:       "warning",
				TargetType: "user",
				IsActive:   false,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("UPDATE notifications").
					WithArgs(
						pgxmock.AnyArg(), // title
						pgxmock.AnyArg(), // body
						pgxmock.AnyArg(), // type
						pgxmock.AnyArg(), // target_type
						pgxmock.AnyArg(), // is_active
						pgxmock.AnyArg(), // id
					).
					WillReturnRows(
						pgxmock.NewRows([]string{"updated_at"}).
							AddRow(time.Now()),
					)
			},
			expectedError: false,
		},
		{
			name: "database error",
			notification: &domain.Notification{
				ID:         uuid.New(),
				Title:      "Fail Update",
				Body:       "Should fail",
				Type:       "info",
				TargetType: "all",
				IsActive:   true,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("UPDATE notifications").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnError(errors.New("update failed"))
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

			err = repo.Update(t.Context(), tt.notification)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.False(t, tt.notification.UpdatedAt.IsZero())
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
