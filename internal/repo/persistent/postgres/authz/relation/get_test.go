package relation

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

	relationID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	tests := []struct {
		name             string
		filter           *domain.RelationFilter
		setupMock        func(pgxmock.PgxPoolIface)
		expectedRelation *domain.Relation
		expectedError    bool
		errorCheck       func(*testing.T, error)
	}{
		{
			name: "success - get by id",
			filter: &domain.RelationFilter{
				ID: &relationID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "type", "name", "created_at"}).
					AddRow(relationID, "group", "engineering-team", now)
				mock.ExpectQuery("SELECT (.+) FROM relation").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedRelation: &domain.Relation{
				ID:        relationID,
				Type:      "group",
				Name:      "engineering-team",
				CreatedAt: now,
			},
			expectedError: false,
		},
		{
			name: "success - get by name",
			filter: &domain.RelationFilter{
				Name: func() *string { s := "engineering-team"; return &s }(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "type", "name", "created_at"}).
					AddRow(relationID, "group", "engineering-team", now)
				mock.ExpectQuery("SELECT (.+) FROM relation").
					WithArgs("engineering-team").
					WillReturnRows(rows)
			},
			expectedRelation: &domain.Relation{
				ID:        relationID,
				Type:      "group",
				Name:      "engineering-team",
				CreatedAt: now,
			},
			expectedError: false,
		},
		{
			name: "success - get by type",
			filter: &domain.RelationFilter{
				Type: func() *string { s := "group"; return &s }(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "type", "name", "created_at"}).
					AddRow(relationID, "group", "engineering-team", now)
				mock.ExpectQuery("SELECT (.+) FROM relation").
					WithArgs("group").
					WillReturnRows(rows)
			},
			expectedRelation: &domain.Relation{
				ID:        relationID,
				Type:      "group",
				Name:      "engineering-team",
				CreatedAt: now,
			},
			expectedError: false,
		},
		{
			name: "success - get by multiple filters",
			filter: &domain.RelationFilter{
				ID:   &relationID,
				Type: func() *string { s := "group"; return &s }(),
				Name: func() *string { s := "engineering-team"; return &s }(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "type", "name", "created_at"}).
					AddRow(relationID, "group", "engineering-team", now)
				mock.ExpectQuery("SELECT (.+) FROM relation").
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedRelation: &domain.Relation{
				ID:        relationID,
				Type:      "group",
				Name:      "engineering-team",
				CreatedAt: now,
			},
			expectedError: false,
		},
		{
			name:   "empty filter",
			filter: &domain.RelationFilter{},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "type", "name", "created_at"}).
					AddRow(relationID, "group", "engineering-team", now)
				mock.ExpectQuery("SELECT (.+) FROM relation").
					WillReturnRows(rows)
			},
			expectedRelation: &domain.Relation{
				ID:        relationID,
				Type:      "group",
				Name:      "engineering-team",
				CreatedAt: now,
			},
			expectedError: false,
		},
		{
			name: "not found",
			filter: &domain.RelationFilter{
				ID: &relationID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM relation").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedRelation: nil,
			expectedError:    true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "database error",
			filter: &domain.RelationFilter{
				ID: &relationID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM relation").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("database error"))
			},
			expectedRelation: nil,
			expectedError:    true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "connection timeout",
			filter: &domain.RelationFilter{
				ID: &relationID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM relation").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("connection timeout"))
			},
			expectedRelation: nil,
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
				assert.Equal(t, tt.expectedRelation.ID, result.ID)
				assert.Equal(t, tt.expectedRelation.Type, result.Type)
				assert.Equal(t, tt.expectedRelation.Name, result.Name)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
