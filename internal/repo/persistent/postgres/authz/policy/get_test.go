package policy

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Get(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	now := time.Now()

	policyID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	permID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440000")
	activeTrue := true

	tests := []struct {
		name           string
		filter         *domain.PolicyFilter
		setupMock      func(pgxmock.PgxPoolIface)
		expectedPolicy *domain.Policy
		expectedError  bool
		errorCheck     func(*testing.T, error)
	}{
		{
			name: "success - get by id",
			filter: &domain.PolicyFilter{
				ID: &policyID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "permission_id", "effect", "priority", "active", "conditions", "created_at",
				}).AddRow(
					policyID, permID, "allow", 10, true, map[string]any{"ip": "10.0.0.0/8"}, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM policy").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedPolicy: &domain.Policy{
				ID:           policyID,
				PermissionID: permID,
				Effect:       "allow",
				Priority:     10,
				Active:       true,
				Conditions:   map[string]any{"ip": "10.0.0.0/8"},
				CreatedAt:    now,
			},
			expectedError: false,
		},
		{
			name: "success - get by permission_id",
			filter: &domain.PolicyFilter{
				PermissionID: &permID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "permission_id", "effect", "priority", "active", "conditions", "created_at",
				}).AddRow(
					policyID, permID, "deny", 5, false, map[string]any{}, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM policy").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedPolicy: &domain.Policy{
				ID:           policyID,
				PermissionID: permID,
				Effect:       "deny",
				Priority:     5,
				Active:       false,
				Conditions:   map[string]any{},
				CreatedAt:    now,
			},
			expectedError: false,
		},
		{
			name: "success - get by active filter",
			filter: &domain.PolicyFilter{
				Active: &activeTrue,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "permission_id", "effect", "priority", "active", "conditions", "created_at",
				}).AddRow(
					policyID, permID, "allow", 10, true, map[string]any{}, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM policy").
					WithArgs(true).
					WillReturnRows(rows)
			},
			expectedPolicy: &domain.Policy{
				ID:           policyID,
				PermissionID: permID,
				Effect:       "allow",
				Priority:     10,
				Active:       true,
				Conditions:   map[string]any{},
				CreatedAt:    now,
			},
			expectedError: false,
		},
		{
			name: "success - get by multiple filters",
			filter: &domain.PolicyFilter{
				ID:           &policyID,
				PermissionID: &permID,
				Active:       &activeTrue,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "permission_id", "effect", "priority", "active", "conditions", "created_at",
				}).AddRow(
					policyID, permID, "allow", 10, true, map[string]any{}, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM policy").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), true).
					WillReturnRows(rows)
			},
			expectedPolicy: &domain.Policy{
				ID:           policyID,
				PermissionID: permID,
				Effect:       "allow",
				Priority:     10,
				Active:       true,
				Conditions:   map[string]any{},
				CreatedAt:    now,
			},
			expectedError: false,
		},
		{
			name:   "empty filter",
			filter: &domain.PolicyFilter{},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "permission_id", "effect", "priority", "active", "conditions", "created_at",
				}).AddRow(
					policyID, permID, "allow", 10, true, map[string]any{}, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM policy").
					WillReturnRows(rows)
			},
			expectedPolicy: &domain.Policy{
				ID:           policyID,
				PermissionID: permID,
				Effect:       "allow",
				Priority:     10,
				Active:       true,
				Conditions:   map[string]any{},
				CreatedAt:    now,
			},
			expectedError: false,
		},
		{
			name: "not found",
			filter: &domain.PolicyFilter{
				ID: &policyID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM policy").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedPolicy: nil,
			expectedError:  true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "database error",
			filter: &domain.PolicyFilter{
				ID: &policyID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM policy").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			expectedPolicy: nil,
			expectedError:  true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "connection timeout",
			filter: &domain.PolicyFilter{
				ID: &policyID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM policy").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("connection timeout"))
			},
			expectedPolicy: nil,
			expectedError:  true,
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

			result, err := repo.Get(ctx, tt.filter)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, result)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedPolicy.ID, result.ID)
				assert.Equal(t, tt.expectedPolicy.PermissionID, result.PermissionID)
				assert.Equal(t, tt.expectedPolicy.Effect, result.Effect)
				assert.Equal(t, tt.expectedPolicy.Priority, result.Priority)
				assert.Equal(t, tt.expectedPolicy.Active, result.Active)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
