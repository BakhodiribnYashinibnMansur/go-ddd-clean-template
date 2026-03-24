package systemerror

import (
	"errors"
	"testing"

	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Resolve(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	id := uuid.New()
	resolvedBy := uuid.New()

	tests := []struct {
		name          string
		id            uuid.UUID
		resolvedBy    *uuid.UUID
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name:       "success",
			id:         id,
			resolvedBy: &resolvedBy,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE system_errors").
					WithArgs(
						pgxmock.AnyArg(), // is_resolved
						pgxmock.AnyArg(), // resolved_at
						pgxmock.AnyArg(), // resolved_by
						pgxmock.AnyArg(), // WHERE id
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: false,
		},
		{
			name:       "success with nil resolved_by",
			id:         id,
			resolvedBy: nil,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE system_errors").
					WithArgs(
						pgxmock.AnyArg(), // is_resolved
						pgxmock.AnyArg(), // resolved_at
						pgxmock.AnyArg(), // resolved_by
						pgxmock.AnyArg(), // WHERE id
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: false,
		},
		{
			name:       "database error",
			id:         id,
			resolvedBy: &resolvedBy,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE system_errors").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnError(errors.New("database error"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name:       "no rows affected",
			id:         uuid.New(),
			resolvedBy: &resolvedBy,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE system_errors").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: false,
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
				log:     logger.New("debug"),
			}

			err = repo.Resolve(ctx, tt.id, tt.resolvedBy)

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
