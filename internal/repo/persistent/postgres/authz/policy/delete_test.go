package policy

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

func TestRepo_Delete(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	policyID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name          string
		id            uuid.UUID
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success",
			id:   policyID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM policy").
					WithArgs(pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			expectedError: false,
		},
		{
			name: "not found - no rows affected",
			id:   policyID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM policy").
					WithArgs(pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			},
		},
		{
			name: "database error",
			id:   policyID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM policy").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "foreign key constraint violation",
			id:   policyID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM policy").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("foreign key constraint violation"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "foreign key")
			},
		},
		{
			name: "nil uuid",
			id:   uuid.Nil,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM policy").
					WithArgs(pgxmock.AnyArg()).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			},
		},
		{
			name: "connection timeout",
			id:   policyID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("DELETE FROM policy").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("connection timeout"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "timeout")
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

			err = repo.Delete(ctx, tt.id)

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
