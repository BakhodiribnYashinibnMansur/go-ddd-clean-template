package policy

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

func TestRepo_GetByRole(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	now := time.Now()

	roleID := uuid.MustParse("770e8400-e29b-41d4-a716-446655440000")
	policyID1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	policyID2 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	permID1 := uuid.MustParse("660e8400-e29b-41d4-a716-446655440000")
	permID2 := uuid.MustParse("660e8400-e29b-41d4-a716-446655440001")

	// The actual SQL: SELECT p.id, ... FROM policy p JOIN role_permission rp ON p.permission_id = rp.permission_id WHERE rp.role_id = $1 AND p.active = $2 ORDER BY p.priority DESC
	sqlRegex := "SELECT (.+) FROM policy p JOIN role_permission rp ON"

	tests := []struct {
		name             string
		roleID           uuid.UUID
		setupMock        func(pgxmock.PgxPoolIface)
		expectedPolicies int
		expectedError    bool
		errorCheck       func(*testing.T, error)
	}{
		{
			name:   "success - multiple policies",
			roleID: roleID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "permission_id", "effect", "priority", "active", "conditions", "created_at",
				}).
					AddRow(policyID1, permID1, "allow", 100, true, map[string]any{"ip": "10.0.0.0/8"}, now).
					AddRow(policyID2, permID2, "deny", 50, true, map[string]any{}, now)

				mock.ExpectQuery(sqlRegex).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedPolicies: 2,
			expectedError:    false,
		},
		{
			name:   "success - single policy",
			roleID: roleID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "permission_id", "effect", "priority", "active", "conditions", "created_at",
				}).AddRow(policyID1, permID1, "allow", 10, true, map[string]any{}, now)

				mock.ExpectQuery(sqlRegex).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedPolicies: 1,
			expectedError:    false,
		},
		{
			name:   "success - empty result",
			roleID: roleID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "permission_id", "effect", "priority", "active", "conditions", "created_at",
				})

				mock.ExpectQuery(sqlRegex).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedPolicies: 0,
			expectedError:    false,
		},
		{
			name:   "database error",
			roleID: roleID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(sqlRegex).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			expectedPolicies: 0,
			expectedError:    true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name:   "connection timeout",
			roleID: roleID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery(sqlRegex).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnError(errors.New("connection timeout"))
			},
			expectedPolicies: 0,
			expectedError:    true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "timeout")
			},
		},
		{
			name:   "nil uuid role",
			roleID: uuid.Nil,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "permission_id", "effect", "priority", "active", "conditions", "created_at",
				})

				mock.ExpectQuery(sqlRegex).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedPolicies: 0,
			expectedError:    false,
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

			policies, err := repo.GetByRole(ctx, tt.roleID)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, policies)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, policies, tt.expectedPolicies)

				// Verify ordering by priority DESC if multiple results
				if len(policies) > 1 {
					assert.GreaterOrEqual(t, policies[0].Priority, policies[1].Priority)
				}
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
