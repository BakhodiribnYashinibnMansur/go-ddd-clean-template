package relation

import (
	"errors"
	"testing"

	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Update(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	relationID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name          string
		relation      *domain.Relation
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success",
			relation: &domain.Relation{
				ID:   relationID,
				Type: "group",
				Name: "updated-engineering-team",
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE relation").
					WithArgs(
						pgxmock.AnyArg(), // name
						pgxmock.AnyArg(), // type
						pgxmock.AnyArg(), // id (WHERE)
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: false,
		},
		{
			name: "success - change type",
			relation: &domain.Relation{
				ID:   relationID,
				Type: "organization",
				Name: "engineering-team",
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE relation").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: false,
		},
		{
			name: "not found - no rows affected",
			relation: &domain.Relation{
				ID:   relationID,
				Type: "group",
				Name: "nonexistent",
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE relation").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			},
		},
		{
			name: "database error",
			relation: &domain.Relation{
				ID:   relationID,
				Type: "group",
				Name: "error-group",
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE relation").
					WithArgs(
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
			name: "duplicate name constraint",
			relation: &domain.Relation{
				ID:   relationID,
				Type: "group",
				Name: "existing-name",
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE relation").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnError(errors.New("duplicate key value violates unique constraint"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "connection timeout",
			relation: &domain.Relation{
				ID:   relationID,
				Type: "group",
				Name: "timeout-group",
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE relation").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
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

			err = repo.Update(ctx, tt.relation)

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
