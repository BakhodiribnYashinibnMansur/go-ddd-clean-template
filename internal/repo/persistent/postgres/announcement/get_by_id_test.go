package announcement

import (
	"errors"
	"testing"
	"time"

	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_GetByID(t *testing.T) {
	ctx := t.Context()
	now := time.Now()
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	startsAt := now.Add(24 * time.Hour)

	tests := []struct {
		name          string
		id            uuid.UUID
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success",
			id:   id,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "title", "content", "type", "is_active",
					"starts_at", "ends_at", "created_at", "updated_at",
				}).AddRow(
					id, "Test Announcement", "Some content", "info", true,
					&startsAt, nil, now, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM announcements").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedError: false,
		},
		{
			name: "not found",
			id:   id,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM announcements").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("no rows in result set"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "database error",
			id:   id,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM announcements").
					WithArgs(pgxmock.AnyArg()).
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

			result, err := repo.GetByID(ctx, tt.id)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, result)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, id, result.ID)
				assert.Equal(t, "Test Announcement", result.Title)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
