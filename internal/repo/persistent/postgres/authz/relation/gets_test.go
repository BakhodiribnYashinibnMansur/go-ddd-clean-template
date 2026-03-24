package relation

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

	relationID1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	relationID2 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")

	tests := []struct {
		name              string
		filter            *domain.RelationsFilter
		setupMock         func(pgxmock.PgxPoolIface)
		expectedRelations int
		expectedCount     int
		expectedError     bool
		errorCheck        func(*testing.T, error)
	}{
		{
			name: "success - get all",
			filter: &domain.RelationsFilter{
				RelationFilter: domain.RelationFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "type", "name", "created_at"}).
					AddRow(relationID1, "group", "engineering", now).
					AddRow(relationID2, "organization", "acme-corp", now)

				mock.ExpectQuery("SELECT (.+) FROM relation").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(2))
			},
			expectedRelations: 2,
			expectedCount:     2,
			expectedError:     false,
		},
		{
			name: "success - filter by type",
			filter: &domain.RelationsFilter{
				RelationFilter: domain.RelationFilter{
					Type: func() *string { s := "group"; return &s }(),
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "type", "name", "created_at"}).
					AddRow(relationID1, "group", "engineering", now)

				mock.ExpectQuery("SELECT (.+) FROM relation").
					WithArgs("group").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WithArgs("group").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
			},
			expectedRelations: 1,
			expectedCount:     1,
			expectedError:     false,
		},
		{
			name: "success - filter by name",
			filter: &domain.RelationsFilter{
				RelationFilter: domain.RelationFilter{
					Name: func() *string { s := "engineering"; return &s }(),
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "type", "name", "created_at"}).
					AddRow(relationID1, "group", "engineering", now)

				mock.ExpectQuery("SELECT (.+) FROM relation").
					WithArgs("engineering").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WithArgs("engineering").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))
			},
			expectedRelations: 1,
			expectedCount:     1,
			expectedError:     false,
		},
		{
			name: "success - with pagination",
			filter: &domain.RelationsFilter{
				RelationFilter: domain.RelationFilter{},
				Pagination:     &domain.Pagination{Limit: 10, Offset: 0},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "type", "name", "created_at"}).
					AddRow(relationID1, "group", "engineering", now)

				mock.ExpectQuery("SELECT (.+) FROM relation").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(25))
			},
			expectedRelations: 1,
			expectedCount:     25,
			expectedError:     false,
		},
		{
			name: "success - empty result",
			filter: &domain.RelationsFilter{
				RelationFilter: domain.RelationFilter{
					ID: func() *uuid.UUID { id := uuid.New(); return &id }(),
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "type", "name", "created_at"})

				mock.ExpectQuery("SELECT (.+) FROM relation").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))
			},
			expectedRelations: 0,
			expectedCount:     0,
			expectedError:     false,
		},
		{
			name: "database error on query",
			filter: &domain.RelationsFilter{
				RelationFilter: domain.RelationFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM relation").
					WillReturnError(errors.New("database error"))
			},
			expectedRelations: 0,
			expectedCount:     0,
			expectedError:     true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "database error on count",
			filter: &domain.RelationsFilter{
				RelationFilter: domain.RelationFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "type", "name", "created_at"}).
					AddRow(relationID1, "group", "engineering", now)

				mock.ExpectQuery("SELECT (.+) FROM relation").
					WillReturnRows(rows)

				mock.ExpectQuery("SELECT COUNT").
					WillReturnError(errors.New("count error"))
			},
			expectedRelations: 0,
			expectedCount:     0,
			expectedError:     true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "connection timeout",
			filter: &domain.RelationsFilter{
				RelationFilter: domain.RelationFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM relation").
					WillReturnError(errors.New("connection timeout"))
			},
			expectedRelations: 0,
			expectedCount:     0,
			expectedError:     true,
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

			relations, count, err := repo.Gets(ctx, tt.filter)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, relations)
				assert.Equal(t, 0, count)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
				assert.Len(t, relations, tt.expectedRelations)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
