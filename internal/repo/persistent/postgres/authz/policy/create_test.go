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

func TestRepo_Create(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	permID := uuid.New()
	now := time.Now()

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
				PermissionID: permID,
				Effect:       "allow",
				Priority:     10,
				Active:       true,
				Conditions:   map[string]any{"ip": "192.168.1.0/24"},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				rows := pgxmock.NewRows([]string{"id", "created_at"}).
					AddRow(uuid.New(), now)
				mock.ExpectQuery("INSERT INTO policy").
					WithArgs(
						pgxmock.AnyArg(), // permission_id
						pgxmock.AnyArg(), // effect
						pgxmock.AnyArg(), // priority
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // conditions
						pgxmock.AnyArg(), // created_at
					).
					WillReturnRows(rows)
				mock.ExpectCommit()
			},
			expectedError: false,
		},
		{
			name: "success - nil conditions",
			policy: &domain.Policy{
				PermissionID: permID,
				Effect:       "deny",
				Priority:     5,
				Active:       false,
				Conditions:   nil,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				rows := pgxmock.NewRows([]string{"id", "created_at"}).
					AddRow(uuid.New(), now)
				mock.ExpectQuery("INSERT INTO policy").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnRows(rows)
				mock.ExpectCommit()
			},
			expectedError: false,
		},
		{
			name: "database error on insert",
			policy: &domain.Policy{
				PermissionID: permID,
				Effect:       "allow",
				Priority:     10,
				Active:       true,
				Conditions:   map[string]any{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery("INSERT INTO policy").
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
				PermissionID: permID,
				Effect:       "allow",
				Priority:     1,
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
				PermissionID: permID,
				Effect:       "allow",
				Priority:     10,
				Active:       true,
				Conditions:   map[string]any{"role": "admin"},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				rows := pgxmock.NewRows([]string{"id", "created_at"}).
					AddRow(uuid.New(), now)
				mock.ExpectQuery("INSERT INTO policy").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnRows(rows)
				mock.ExpectCommit().WillReturnError(errors.New("commit error"))
				mock.ExpectRollback()
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "duplicate policy error",
			policy: &domain.Policy{
				PermissionID: permID,
				Effect:       "allow",
				Priority:     10,
				Active:       true,
				Conditions:   map[string]any{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				mock.ExpectQuery("INSERT INTO policy").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnError(errors.New("duplicate key value violates unique constraint"))
				mock.ExpectRollback()
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "success - complex conditions",
			policy: &domain.Policy{
				PermissionID: permID,
				Effect:       "allow",
				Priority:     100,
				Active:       true,
				Conditions: map[string]any{
					"ip_range":   "10.0.0.0/8",
					"time_after":  "09:00",
					"time_before": "17:00",
					"departments": []string{"engineering", "product"},
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectBegin()
				rows := pgxmock.NewRows([]string{"id", "created_at"}).
					AddRow(uuid.New(), now)
				mock.ExpectQuery("INSERT INTO policy").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnRows(rows)
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

			err = repo.Create(ctx, tt.policy)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, tt.policy.ID)
				assert.False(t, tt.policy.CreatedAt.IsZero())
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
