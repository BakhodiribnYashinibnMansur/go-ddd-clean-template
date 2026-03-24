package featureflag

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

	ctx := t.Context()
	now := time.Now()
	id1 := uuid.New()
	id2 := uuid.New()

	tests := []struct {
		name          string
		filter        domain.FeatureFlagFilter
		setupMock     func(pgxmock.PgxPoolIface)
		expectedLen   int
		expectedTotal int64
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name:   "success with items",
			filter: domain.FeatureFlagFilter{Limit: 10, Offset: 0},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(2))
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{
					"id", "key", "name", "type", "value", "description", "is_active",
					"created_at", "updated_at",
				}).
					AddRow(id1, "flag_1", "Flag One", "boolean", "true", "First flag", true, now, now).
					AddRow(id2, "flag_2", "Flag Two", "string", "hello", "Second flag", false, now, now)
				mock.ExpectQuery("SELECT (.+) FROM feature_flags").
					WillReturnRows(rows)
			},
			expectedLen:   2,
			expectedTotal: 2,
			expectedError: false,
		},
		{
			name:   "empty list",
			filter: domain.FeatureFlagFilter{Limit: 10, Offset: 0},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(0))
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{
					"id", "key", "name", "type", "value", "description", "is_active",
					"created_at", "updated_at",
				})
				mock.ExpectQuery("SELECT (.+) FROM feature_flags").
					WillReturnRows(rows)
			},
			expectedLen:   0,
			expectedTotal: 0,
			expectedError: false,
		},
		{
			name:   "count query error",
			filter: domain.FeatureFlagFilter{Limit: 10},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT COUNT").
					WillReturnError(errors.New("database error"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name:   "list query error",
			filter: domain.FeatureFlagFilter{Limit: 10},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(countRows)

				mock.ExpectQuery("SELECT (.+) FROM feature_flags").
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

			items, total, err := repo.List(ctx, tt.filter)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, items, tt.expectedLen)
				assert.Equal(t, tt.expectedTotal, total)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
