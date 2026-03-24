package policy

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

func TestRepo_Gets(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	now := time.Now()

	policyID1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	policyID2 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	permID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440000")
	activeTrue := true
	activeFalse := false

	tests := []struct {
		name             string
		filter           *domain.PoliciesFilter
		setupMock        func(pgxmock.PgxPoolIface)
		expectedPolicies int
		expectedCount    int
		expectedError    bool
		errorCheck       func(*testing.T, error)
	}{
		{
			name: "success - get all",
			filter: &domain.PoliciesFilter{
				PolicyFilter: domain.PolicyFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "permission_id", "effect", "priority", "active", "conditions", "created_at",
				}).
					AddRow(policyID1, permID, "allow", 10, true, map[string]any{}, now).
					AddRow(policyID2, permID, "deny", 5, false, map[string]any{}, now)

				mock.ExpectQuery("SELECT (.+) FROM policy").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(2))
			},
			expectedPolicies: 2,
			expectedCount:    2,
			expectedError:    false,
		},
		{
			name: "success - filter by active true",
			filter: &domain.PoliciesFilter{
				PolicyFilter: domain.PolicyFilter{
					Active: &activeTrue,
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "permission_id", "effect", "priority", "active", "conditions", "created_at",
				}).AddRow(policyID1, permID, "allow", 10, true, map[string]any{}, now)

				mock.ExpectQuery("SELECT (.+) FROM policy").
					WithArgs(true).
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WithArgs(true).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
			},
			expectedPolicies: 1,
			expectedCount:    1,
			expectedError:    false,
		},
		{
			name: "success - filter by active false",
			filter: &domain.PoliciesFilter{
				PolicyFilter: domain.PolicyFilter{
					Active: &activeFalse,
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "permission_id", "effect", "priority", "active", "conditions", "created_at",
				}).AddRow(policyID2, permID, "deny", 5, false, map[string]any{}, now)

				mock.ExpectQuery("SELECT (.+) FROM policy").
					WithArgs(false).
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WithArgs(false).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
			},
			expectedPolicies: 1,
			expectedCount:    1,
			expectedError:    false,
		},
		{
			name: "success - filter by permission_id",
			filter: &domain.PoliciesFilter{
				PolicyFilter: domain.PolicyFilter{
					PermissionID: &permID,
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "permission_id", "effect", "priority", "active", "conditions", "created_at",
				}).AddRow(policyID1, permID, "allow", 10, true, map[string]any{}, now)

				mock.ExpectQuery("SELECT (.+) FROM policy").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
			},
			expectedPolicies: 1,
			expectedCount:    1,
			expectedError:    false,
		},
		{
			name: "success - with pagination",
			filter: &domain.PoliciesFilter{
				PolicyFilter: domain.PolicyFilter{},
				Pagination:   &domain.Pagination{Limit: 10, Offset: 0},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "permission_id", "effect", "priority", "active", "conditions", "created_at",
				}).AddRow(policyID1, permID, "allow", 10, true, map[string]any{}, now)

				mock.ExpectQuery("SELECT (.+) FROM policy").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(25))
			},
			expectedPolicies: 1,
			expectedCount:    25,
			expectedError:    false,
		},
		{
			name: "success - empty result",
			filter: &domain.PoliciesFilter{
				PolicyFilter: domain.PolicyFilter{
					ID: func() *uuid.UUID { id := uuid.New(); return &id }(),
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "permission_id", "effect", "priority", "active", "conditions", "created_at",
				})

				mock.ExpectQuery("SELECT (.+) FROM policy").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedPolicies: 0,
			expectedCount:    0,
			expectedError:    false,
		},
		{
			name: "database error on query",
			filter: &domain.PoliciesFilter{
				PolicyFilter: domain.PolicyFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM policy").
					WillReturnError(errors.New("database error"))
			},
			expectedPolicies: 0,
			expectedCount:    0,
			expectedError:    true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "database error on count",
			filter: &domain.PoliciesFilter{
				PolicyFilter: domain.PolicyFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "permission_id", "effect", "priority", "active", "conditions", "created_at",
				}).AddRow(policyID1, permID, "allow", 10, true, map[string]any{}, now)

				mock.ExpectQuery("SELECT (.+) FROM policy").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WillReturnError(errors.New("count error"))
			},
			expectedPolicies: 0,
			expectedCount:    0,
			expectedError:    true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "connection timeout",
			filter: &domain.PoliciesFilter{
				PolicyFilter: domain.PolicyFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM policy").
					WillReturnError(errors.New("connection timeout"))
			},
			expectedPolicies: 0,
			expectedCount:    0,
			expectedError:    true,
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

			policies, count, err := repo.Gets(ctx, tt.filter)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, policies)
				assert.Equal(t, 0, count)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
				assert.Len(t, policies, tt.expectedPolicies)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
