package policy

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

	policyID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	permID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name          string
		policy        *domain.Policy
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success",
			policy: &domain.Policy{
				ID:           policyID,
				PermissionID: permID,
				Effect:       "allow",
				Priority:     10,
				Active:       true,
				Conditions:   map[string]any{"ip": "10.0.0.0/8"},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE policy").
					WithArgs(
						pgxmock.AnyArg(), // permission_id
						pgxmock.AnyArg(), // effect
						pgxmock.AnyArg(), // priority
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // conditions
						pgxmock.AnyArg(), // id (WHERE)
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectCommit()
			},
			expectedError: false,
		},
		{
			name: "success - nil conditions",
			policy: &domain.Policy{
				ID:           policyID,
				PermissionID: permID,
				Effect:       "deny",
				Priority:     5,
				Active:       false,
				Conditions:   nil,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE policy").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectCommit()
			},
			expectedError: false,
		},
		{
			name: "not found - no rows affected",
			policy: &domain.Policy{
				ID:           policyID,
				PermissionID: permID,
				Effect:       "allow",
				Priority:     10,
				Active:       true,
				Conditions:   map[string]any{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE policy").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
				mock.ExpectRollback()
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			},
		},
		{
			name: "database error",
			policy: &domain.Policy{
				ID:           policyID,
				PermissionID: permID,
				Effect:       "allow",
				Priority:     10,
				Active:       true,
				Conditions:   map[string]any{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE policy").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnError(errors.New("database error"))
				mock.ExpectRollback()
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "begin transaction error",
			policy: &domain.Policy{
				ID:           policyID,
				PermissionID: permID,
				Effect:       "allow",
				Priority:     10,
				Active:       true,
				Conditions:   map[string]any{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin().WillReturnError(errors.New("begin error"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "commit error",
			policy: &domain.Policy{
				ID:           policyID,
				PermissionID: permID,
				Effect:       "allow",
				Priority:     10,
				Active:       true,
				Conditions:   map[string]any{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE policy").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectCommit().WillReturnError(errors.New("commit error"))
				mock.ExpectRollback()
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "success - complex conditions update",
			policy: &domain.Policy{
				ID:           policyID,
				PermissionID: permID,
				Effect:       "allow",
				Priority:     100,
				Active:       true,
				Conditions: map[string]any{
					"ip_range":    "10.0.0.0/8",
					"departments": []string{"engineering"},
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectExec("UPDATE policy").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
				mock.ExpectCommit()
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
				logger:  logger.New("debug"),
			}

			err = repo.Update(ctx, tt.policy)

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
