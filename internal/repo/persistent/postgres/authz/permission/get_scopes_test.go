package permission

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

func TestRepo_GetScopes(t *testing.T) {
	ctx := t.Context()
	now := time.Now()

	permID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name           string
		permID         uuid.UUID
		setupMock      func(pgxmock.PgxPoolIface)
		expectedScopes []*domain.Scope
		expectedError  bool
		errorCheck     func(*testing.T, error)
	}{
		{
			name:   "success - multiple scopes",
			permID: permID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"path", "method", "created_at"}).
					AddRow("/api/v1/users", "GET", now).
					AddRow("/api/v1/users", "POST", now)

				mock.ExpectQuery("SELECT (.+) FROM scope").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedScopes: []*domain.Scope{
				{Path: "/api/v1/users", Method: "GET", CreatedAt: now},
				{Path: "/api/v1/users", Method: "POST", CreatedAt: now},
			},
			expectedError: false,
		},
		{
			name:   "success - empty result",
			permID: permID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"path", "method", "created_at"})

				mock.ExpectQuery("SELECT (.+) FROM scope").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedScopes: nil,
			expectedError:  false,
		},
		{
			name:   "database error",
			permID: permID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM scope").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			expectedScopes: nil,
			expectedError:  true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name:   "success - single scope",
			permID: permID,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"path", "method", "created_at"}).
					AddRow("/api/v1/roles", "DELETE", now)

				mock.ExpectQuery("SELECT (.+) FROM scope").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedScopes: []*domain.Scope{
				{Path: "/api/v1/roles", Method: "DELETE", CreatedAt: now},
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

			scopes, err := repo.GetScopes(ctx, tt.permID)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, scopes)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, len(tt.expectedScopes), len(scopes))
				for i, expected := range tt.expectedScopes {
					assert.Equal(t, expected.Path, scopes[i].Path)
					assert.Equal(t, expected.Method, scopes[i].Method)
				}
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
