package announcement

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
	ctx := t.Context()
	now := time.Now()
	startsAt := now.Add(24 * time.Hour)

	ann := &domain.Announcement{
		ID:       uuid.New(),
		Title:    "Test Announcement",
		Content:  "Test content here",
		Type:     "info",
		IsActive: true,
		StartsAt: &startsAt,
		EndsAt:   nil,
	}

	tests := []struct {
		name          string
		announcement  *domain.Announcement
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name:         "success",
			announcement: ann,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"created_at", "updated_at"}).
					AddRow(now, now)
				mock.ExpectQuery("INSERT INTO announcements").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // title
						pgxmock.AnyArg(), // content
						pgxmock.AnyArg(), // type
						pgxmock.AnyArg(), // is_active
						pgxmock.AnyArg(), // starts_at
						pgxmock.AnyArg(), // ends_at
					).
					WillReturnRows(rows)
			},
			expectedError: false,
		},
		{
			name:         "database error",
			announcement: ann,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("INSERT INTO announcements").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // title
						pgxmock.AnyArg(), // content
						pgxmock.AnyArg(), // type
						pgxmock.AnyArg(), // is_active
						pgxmock.AnyArg(), // starts_at
						pgxmock.AnyArg(), // ends_at
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
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			tt.setupMock(mockPool)

			repo := &Repo{
				pool:    mockPool,
				builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
				logger:  logger.New("debug"),
			}

			err = repo.Create(ctx, tt.announcement)

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
