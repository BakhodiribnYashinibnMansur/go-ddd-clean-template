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

func TestRepo_Update(t *testing.T) {
	ctx := t.Context()
	now := time.Now()
	updatedAt := now.Add(time.Second)

	ann := &domain.Announcement{
		ID:       uuid.New(),
		Title:    "Updated Title",
		Content:  "Updated Content",
		Type:     "warning",
		IsActive: false,
		StartsAt: nil,
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
				rows := pgxmock.NewRows([]string{"updated_at"}).AddRow(updatedAt)
				mock.ExpectQuery("UPDATE announcements").
					WithArgs(
						pgxmock.AnyArg(), // title
						pgxmock.AnyArg(), // content
						pgxmock.AnyArg(), // type
						pgxmock.AnyArg(), // is_active
						pgxmock.AnyArg(), // starts_at
						pgxmock.AnyArg(), // ends_at
						pgxmock.AnyArg(), // id (WHERE)
					).
					WillReturnRows(rows)
			},
			expectedError: false,
		},
		{
			name:         "database error",
			announcement: ann,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("UPDATE announcements").
					WithArgs(
						pgxmock.AnyArg(), // title
						pgxmock.AnyArg(), // content
						pgxmock.AnyArg(), // type
						pgxmock.AnyArg(), // is_active
						pgxmock.AnyArg(), // starts_at
						pgxmock.AnyArg(), // ends_at
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
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			tt.setupMock(mockPool)

			repo := &Repo{
				pool:    mockPool,
				builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
				logger:  logger.New("debug"),
			}

			err = repo.Update(ctx, tt.announcement)

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
