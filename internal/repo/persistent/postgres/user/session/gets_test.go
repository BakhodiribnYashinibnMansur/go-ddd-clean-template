package session

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"
	"gct/pkg/logger"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Gets(t *testing.T) {
	ctx := t.Context()
	now := time.Now()

	sID1 := uuid.New()
	sID2 := uuid.New()
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	revoked := false

	tests := []struct {
		name             string
		filter           *domain.SessionsFilter
		setupMock        func(pgxmock.PgxPoolIface)
		expectedSessions int
		expectedCount    int
		expectedError    bool
	}{
		{
			name: "success with no filters",
			filter: &domain.SessionsFilter{
				SessionFilter: domain.SessionFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				// Count query
				mock.ExpectQuery("SELECT COUNT(.+) FROM session").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(2))

				// Select query
				rows := pgxmock.NewRows([]string{
					"id", "device_id", "device_name", "device_type", "ip_address", "user_agent",
					"fcm_token", "refresh_token_hash", "data", "user_id", "expires_at",
					"last_activity", "revoked", "created_at", "updated_at",
				}).
					AddRow(sID1, uuid.New(), nil, nil, nil, nil, nil, "hash", []byte("{}"), userID, now, now, false, now, now).
					AddRow(sID2, uuid.New(), nil, nil, nil, nil, nil, "hash", []byte("{}"), userID, now, now, false, now, now)

				mock.ExpectQuery("SELECT (.+) FROM session").
					WillReturnRows(rows)
			},
			expectedSessions: 2,
			expectedCount:    2,
			expectedError:    false,
		},
		{
			name: "success with filters",
			filter: &domain.SessionsFilter{
				SessionFilter: domain.SessionFilter{
					UserID:  &userID,
					Revoked: &revoked,
				},
				Pagination: &domain.Pagination{
					Limit:  10,
					Offset: 0,
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				// Count query
				mock.ExpectQuery("SELECT COUNT(.+) FROM session").
					WithArgs(pgxmock.AnyArg(), revoked).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))

				// Select query
				rows := pgxmock.NewRows([]string{
					"id", "device_id", "device_name", "device_type", "ip_address", "user_agent",
					"fcm_token", "refresh_token_hash", "data", "user_id", "expires_at",
					"last_activity", "revoked", "created_at", "updated_at",
				}).
					AddRow(sID1, uuid.New(), nil, nil, nil, nil, nil, "hash", []byte("{}"), userID, now, now, false, now, now)

				mock.ExpectQuery("SELECT (.+) FROM session").
					WithArgs(pgxmock.AnyArg(), revoked).
					WillReturnRows(rows)
			},
			expectedSessions: 1,
			expectedCount:    1,
			expectedError:    false,
		},
		{
			name: "count error",
			filter: &domain.SessionsFilter{
				SessionFilter: domain.SessionFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT COUNT(.+) FROM session").
					WillReturnError(errors.New("count error"))
			},
			expectedSessions: 0,
			expectedCount:    0,
			expectedError:    true,
		},
		{
			name: "select error",
			filter: &domain.SessionsFilter{
				SessionFilter: domain.SessionFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT COUNT(.+) FROM session").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))

				mock.ExpectQuery("SELECT (.+) FROM session").
					WillReturnError(errors.New("select error"))
			},
			expectedSessions: 0,
			expectedCount:    0,
			expectedError:    true,
		},
		{
			name: "success with pagination",
			filter: &domain.SessionsFilter{
				SessionFilter: domain.SessionFilter{
					UserID: &userID,
				},
				Pagination: &domain.Pagination{
					Limit:  5,
					Offset: 10,
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				// Count query
				mock.ExpectQuery("SELECT COUNT(.+) FROM session").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(25))

				// Select query
				rows := pgxmock.NewRows([]string{
					"id", "device_id", "device_name", "device_type", "ip_address", "user_agent",
					"fcm_token", "refresh_token_hash", "data", "user_id", "expires_at",
					"last_activity", "revoked", "created_at", "updated_at",
				}).AddRow(
					sID1, uuid.New(), nil, nil, nil, nil,
					nil, "hash", []byte("{}"), userID, now, now, false, now, now,
				)

				mock.ExpectQuery("SELECT (.+) FROM session").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedSessions: 1,
			expectedCount:    25,
			expectedError:    false,
		},
		{
			name: "success with revoked filter",
			filter: &domain.SessionsFilter{
				SessionFilter: domain.SessionFilter{
					Revoked: func() *bool { r := true; return &r }(),
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				// Count query
				mock.ExpectQuery("SELECT COUNT(.+) FROM session").
					WithArgs(true).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(3))

				// Select query
				rows := pgxmock.NewRows([]string{
					"id", "device_id", "device_name", "device_type", "ip_address", "user_agent",
					"fcm_token", "refresh_token_hash", "data", "user_id", "expires_at",
					"last_activity", "revoked", "created_at", "updated_at",
				}).AddRow(
					sID1, uuid.New(), nil, nil, nil, nil,
					nil, "hash", []byte("{}"), userID, now, now, true, now, now,
				)

				mock.ExpectQuery("SELECT (.+) FROM session").
					WithArgs(true).
					WillReturnRows(rows)
			},
			expectedSessions: 1,
			expectedCount:    3,
			expectedError:    false,
		},
		{
			name: "connection timeout on count",
			filter: &domain.SessionsFilter{
				SessionFilter: domain.SessionFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT COUNT(.+) FROM session").
					WillReturnError(errors.New("connection timeout"))
			},
			expectedSessions: 0,
			expectedCount:    0,
			expectedError:    true,
		},
		{
			name: "empty result with filters",
			filter: &domain.SessionsFilter{
				SessionFilter: domain.SessionFilter{
					UserID: func() *uuid.UUID { id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"); return &id }(),
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				// Count query
				mock.ExpectQuery("SELECT COUNT(.+) FROM session").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))

				// Select query
				rows := pgxmock.NewRows([]string{
					"id", "device_id", "device_name", "device_type", "ip_address", "user_agent",
					"fcm_token", "refresh_token_hash", "data", "user_id", "expires_at",
					"last_activity", "revoked", "created_at", "updated_at",
				})

				mock.ExpectQuery("SELECT (.+) FROM session").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedSessions: 0,
			expectedCount:    0,
			expectedError:    false,
		},
		{
			name: "success with multiple filters",
			filter: &domain.SessionsFilter{
				SessionFilter: domain.SessionFilter{
					UserID:  &userID,
					Revoked: &revoked,
				},
				Pagination: &domain.Pagination{
					Limit:  20,
					Offset: 0,
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				// Count query
				mock.ExpectQuery("SELECT COUNT(.+) FROM session").
					WithArgs(pgxmock.AnyArg(), revoked).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(5))

				// Select query
				rows := pgxmock.NewRows([]string{
					"id", "device_id", "device_name", "device_type", "ip_address", "user_agent",
					"fcm_token", "refresh_token_hash", "data", "user_id", "expires_at",
					"last_activity", "revoked", "created_at", "updated_at",
				}).AddRow(
					sID1, uuid.New(), nil, nil, nil, nil,
					nil, "hash", []byte("{}"), userID, now, now, false, now, now,
				)

				mock.ExpectQuery("SELECT (.+) FROM session").
					WithArgs(pgxmock.AnyArg(), revoked).
					WillReturnRows(rows)
			},
			expectedSessions: 1,
			expectedCount:    5,
			expectedError:    false,
		},
		{
			name: "sessions with nil optional fields",
			filter: &domain.SessionsFilter{
				SessionFilter: domain.SessionFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				// Count query
				mock.ExpectQuery("SELECT COUNT(.+) FROM session").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))

				// Select query
				rows := pgxmock.NewRows([]string{
					"id", "device_id", "device_name", "device_type", "ip_address", "user_agent",
					"fcm_token", "refresh_token_hash", "data", "user_id", "expires_at",
					"last_activity", "revoked", "created_at", "updated_at",
				}).AddRow(
					sID1, uuid.New(), nil, nil, nil, nil,
					nil, "hash", []byte("{}"), userID, now, now, false, now, now,
				)

				mock.ExpectQuery("SELECT (.+) FROM session").
					WillReturnRows(rows)
			},
			expectedSessions: 1,
			expectedCount:    1,
			expectedError:    false,
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

			sessions, count, err := repo.Gets(ctx, tt.filter)
			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, sessions, tt.expectedSessions)
				assert.Equal(t, tt.expectedCount, count)
			}
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
