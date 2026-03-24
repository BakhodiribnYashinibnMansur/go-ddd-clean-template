package job

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

	j := &domain.Job{
		ID:           uuid.New(),
		Name:         "Email Sender",
		Type:         "cron",
		CronSchedule: "0 * * * *",
		Payload:      map[string]any{"template": "welcome"},
		IsActive:     true,
		Status:       "idle",
	}

	tests := []struct {
		name          string
		job           *domain.Job
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success",
			job:  j,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"created_at", "updated_at"}).
					AddRow(now, now)
				mock.ExpectQuery("INSERT INTO jobs").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // name
						pgxmock.AnyArg(), // type
						pgxmock.AnyArg(), // cron_schedule
						pgxmock.AnyArg(), // payload (JSON)
						pgxmock.AnyArg(), // is_active
						pgxmock.AnyArg(), // status
					).
					WillReturnRows(rows)
			},
			expectedError: false,
		},
		{
			name: "database error",
			job:  j,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("INSERT INTO jobs").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // name
						pgxmock.AnyArg(), // type
						pgxmock.AnyArg(), // cron_schedule
						pgxmock.AnyArg(), // payload (JSON)
						pgxmock.AnyArg(), // is_active
						pgxmock.AnyArg(), // status
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

			err = repo.Create(ctx, tt.job)

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
