package session

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
	sID := uuid.New()
	fcm := "new_fcm"

	s := &domain.Session{
		ID:           sID,
		FCMToken:     &fcm,
		Revoked:      true,
		LastActivity: time.Now(),
	}

	tests := []struct {
		name          string
		input         *domain.Session
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
	}{
		{
			name:  "success",
			input: s,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE session").
					WithArgs(
						pgxmock.AnyArg(), // fcm_token
						pgxmock.AnyArg(), // revoked
						pgxmock.AnyArg(), // last_activity
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // where id
					).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: false,
		},
		{
			name:  "db error",
			input: s,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).WillReturnError(errors.New("db error"))
			},
			expectedError: true,
		},
		{
			name: "update with nil FCM token",
			input: &domain.Session{
				ID:           sID,
				FCMToken:     nil,
				Revoked:      true,
				LastActivity: time.Now(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: false,
		},
		{
			name: "update with empty FCM token",
			input: &domain.Session{
				ID:           sID,
				FCMToken:     func() *string { s := ""; return &s }(),
				Revoked:      false,
				LastActivity: time.Now(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: false,
		},
		{
			name:  "connection timeout",
			input: s,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).WillReturnError(errors.New("connection timeout"))
			},
			expectedError: true,
		},
		{
			name:  "session not found",
			input: s,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: false,
		},
		{
			name:  "foreign key constraint violation",
			input: s,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).WillReturnError(errors.New("foreign key constraint violation"))
			},
			expectedError: true,
		},
		{
			name: "update only FCM token",
			input: &domain.Session{
				ID:       sID,
				FCMToken: &fcm,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: false,
		},
		{
			name: "update only revoked status",
			input: &domain.Session{
				ID:      sID,
				Revoked: false,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: false,
		},
		{
			name: "update only last activity",
			input: &domain.Session{
				ID:           sID,
				LastActivity: time.Now().Add(-1 * time.Hour),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: false,
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

			err = repo.Update(ctx, tt.input)
			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestRepo_Revoke(t *testing.T) {
	ctx := t.Context()
	sID := uuid.New()
	filter := &domain.SessionFilter{ID: &sID}

	tests := []struct {
		name          string
		filter        *domain.SessionFilter
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
	}{
		{
			name:   "success",
			filter: filter,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE session").
					WithArgs(
						true,             // revoked
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // where id
					).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: false,
		},
		{
			name:   "db error",
			filter: filter,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE session").
					WithArgs(
						true,
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).WillReturnError(errors.New("db error"))
			},
			expectedError: true,
		},
		{
			name:   "connection timeout",
			filter: filter,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE session").
					WithArgs(
						true,
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).WillReturnError(errors.New("connection timeout"))
			},
			expectedError: true,
		},
		{
			name:   "session not found",
			filter: filter,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE session").
					WithArgs(
						true,
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: false,
		},
		{
			name:   "foreign key constraint violation",
			filter: filter,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE session").
					WithArgs(
						true,
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).WillReturnError(errors.New("foreign key constraint violation"))
			},
			expectedError: true,
		},
		{
			name:   "permission denied",
			filter: filter,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE session").
					WithArgs(
						true,
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).WillReturnError(errors.New("permission denied for table session"))
			},
			expectedError: true,
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

			err = repo.Revoke(ctx, tt.filter)
			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
