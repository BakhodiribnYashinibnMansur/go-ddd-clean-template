package session

import (
	"errors"
	"testing"

	"gct/internal/domain"
	"gct/pkg/logger"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Delete(t *testing.T) {
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
				mock.ExpectExec("DELETE FROM session").
					WithArgs(pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			expectedError: false,
		},
		{
			name:   "db error",
			filter: filter,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM session").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("db error"))
			},
			expectedError: true,
		},
		{
			name: "delete by user ID",
			filter: &domain.SessionFilter{
				UserID: func() *uuid.UUID { id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"); return &id }(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM session").
					WithArgs(pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			expectedError: false,
		},
		{
			name: "delete revoked sessions",
			filter: &domain.SessionFilter{
				Revoked: func() *bool { r := true; return &r }(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM session").
					WithArgs(pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("DELETE", 5))
			},
			expectedError: false,
		},
		{
			name: "delete non-existent session",
			filter: &domain.SessionFilter{
				ID: func() *uuid.UUID { id := uuid.New(); return &id }(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM session").
					WithArgs(pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			expectedError: false,
		},
		{
			name:   "connection timeout",
			filter: filter,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM session").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("connection timeout"))
			},
			expectedError: true,
		},
		{
			name:   "foreign key constraint violation",
			filter: filter,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM session").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("foreign key constraint violation"))
			},
			expectedError: true,
		},
		{
			name:   "empty filter",
			filter: &domain.SessionFilter{},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM session").
					WillReturnResult(pgxmock.NewResult("DELETE", 10))
			},
			expectedError: false,
		},
		{
			name: "delete by user ID and revoked",
			filter: &domain.SessionFilter{
				UserID:  func() *uuid.UUID { id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"); return &id }(),
				Revoked: func() *bool { r := false; return &r }(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM session").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("DELETE", 3))
			},
			expectedError: false,
		},
		{
			name:   "permission denied",
			filter: filter,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM session").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("permission denied for table session"))
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

			err = repo.Delete(ctx, tt.filter)
			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
