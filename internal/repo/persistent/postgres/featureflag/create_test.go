package featureflag

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
	now := time.Now()

	ff := &domain.FeatureFlag{
		ID:          uuid.New(),
		Key:         "enable_dark_mode",
		Name:        "Dark Mode",
		Type:        "boolean",
		Value:       "true",
		Description: "Enable dark mode for users",
		IsActive:    true,
	}

	tests := []struct {
		name          string
		featureFlag   *domain.FeatureFlag
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name:        "success",
			featureFlag: ff,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"created_at", "updated_at"}).
					AddRow(now, now)
				mock.ExpectQuery("INSERT INTO feature_flags").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // key
						pgxmock.AnyArg(), // name
						pgxmock.AnyArg(), // type
						pgxmock.AnyArg(), // value
						pgxmock.AnyArg(), // description
						pgxmock.AnyArg(), // is_active
					).
					WillReturnRows(rows)
			},
			expectedError: false,
		},
		{
			name:        "database error",
			featureFlag: ff,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("INSERT INTO feature_flags").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // key
						pgxmock.AnyArg(), // name
						pgxmock.AnyArg(), // type
						pgxmock.AnyArg(), // value
						pgxmock.AnyArg(), // description
						pgxmock.AnyArg(), // is_active
					).
					WillReturnError(errors.New("database error"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
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

			err = repo.Create(ctx, tt.featureFlag)

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
