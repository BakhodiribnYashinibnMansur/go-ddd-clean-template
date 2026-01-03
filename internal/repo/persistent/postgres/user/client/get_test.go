package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gct/internal/domain"
	"gct/pkg/logger"
)

func TestRepo_Get(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	phone := "+998901234567"
	username := "testuser"
	passwordHash := "hashed_password"
	salt := "salt123"

	tests := []struct {
		name          string
		filter        *domain.UserFilter
		setupMock     func(pgxmock.PgxPoolIface)
		expectedUser  *domain.User
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success - get by id",
			filter: &domain.UserFilter{
				ID: &userID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "username", "phone", "password_hash", "salt",
					"created_at", "updated_at", "deleted_at", "last_seen",
				}).AddRow(
					userID, username, phone, passwordHash, salt,
					now, now, 0, now,
				)

				mock.ExpectQuery("SELECT (.+) FROM users").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expectedUser: &domain.User{
				ID:           userID,
				Username:     &username,
				Phone:        &phone,
				PasswordHash: passwordHash,
				Salt:         &salt,
				CreatedAt:    now,
				UpdatedAt:    now,
				DeletedAt:    0,
				LastSeen:     &now,
			},
			expectedError: false,
		},
		{
			name: "success - get by phone",
			filter: &domain.UserFilter{
				Phone: &phone,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "username", "phone", "password_hash", "salt",
					"created_at", "updated_at", "deleted_at", "last_seen",
				}).AddRow(
					userID, username, phone, passwordHash, salt,
					now, now, 0, now,
				)

				mock.ExpectQuery("SELECT (.+) FROM users").
					WithArgs(phone).
					WillReturnRows(rows)
			},
			expectedUser: &domain.User{
				ID:           userID,
				Username:     &username,
				Phone:        &phone,
				PasswordHash: passwordHash,
				Salt:         &salt,
				CreatedAt:    now,
				UpdatedAt:    now,
				DeletedAt:    0,
				LastSeen:     &now,
			},
			expectedError: false,
		},
		{
			name: "user not found",
			filter: &domain.UserFilter{
				ID: &userID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM users").
					WithArgs(userID).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedUser:  nil,
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.NotNil(t, err)
			},
		},
		{
			name: "database error",
			filter: &domain.UserFilter{
				ID: &userID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM users").
					WithArgs(userID).
					WillReturnError(errors.New("database error"))
			},
			expectedUser:  nil,
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.NotNil(t, err)
			},
		},
		{
			name:   "empty filter",
			filter: &domain.UserFilter{},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "username", "phone", "password_hash", "salt",
					"created_at", "updated_at", "deleted_at", "last_seen",
				}).AddRow(
					userID, username, phone, passwordHash, salt,
					now, now, 0, now,
				)

				mock.ExpectQuery("SELECT (.+) FROM users").
					WillReturnRows(rows)
			},
			expectedUser: &domain.User{
				ID:           userID,
				Username:     &username,
				Phone:        &phone,
				PasswordHash: passwordHash,
				Salt:         &salt,
				CreatedAt:    now,
				UpdatedAt:    now,
				DeletedAt:    0,
				LastSeen:     &now,
			},
			expectedError: false,
		},
		{
			name: "success - get by id and phone",
			filter: &domain.UserFilter{
				ID:    &userID,
				Phone: &phone,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "username", "phone", "password_hash", "salt",
					"created_at", "updated_at", "deleted_at", "last_seen",
				}).AddRow(
					userID, username, phone, passwordHash, salt,
					now, now, 0, now,
				)

				mock.ExpectQuery("SELECT (.+) FROM users").
					WithArgs(userID, phone).
					WillReturnRows(rows)
			},
			expectedUser: &domain.User{
				ID:           userID,
				Username:     &username,
				Phone:        &phone,
				PasswordHash: passwordHash,
				Salt:         &salt,
				CreatedAt:    now,
				UpdatedAt:    now,
				DeletedAt:    0,
				LastSeen:     &now,
			},
			expectedError: false,
		},
		{
			name: "user with nil fields",
			filter: &domain.UserFilter{
				ID: &userID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "username", "phone", "password_hash", "salt",
					"created_at", "updated_at", "deleted_at", "last_seen",
				}).AddRow(
					userID, nil, phone, passwordHash, nil,
					now, now, 0, nil,
				)

				mock.ExpectQuery("SELECT (.+) FROM users").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expectedUser: &domain.User{
				ID:           userID,
				Username:     nil,
				Phone:        &phone,
				PasswordHash: passwordHash,
				Salt:         nil,
				CreatedAt:    now,
				UpdatedAt:    now,
				DeletedAt:    0,
				LastSeen:     nil,
			},
			expectedError: false,
		},
		{
			name: "connection timeout",
			filter: &domain.UserFilter{
				ID: &userID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM users").
					WithArgs(userID).
					WillReturnError(errors.New("connection timeout"))
			},
			expectedUser:  nil,
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "timeout")
			},
		},
		{
			name: "invalid phone format in filter",
			filter: &domain.UserFilter{
				Phone: func() *string { s := "invalid-phone"; return &s }(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM users").
					WithArgs("invalid-phone").
					WillReturnError(errors.New("invalid phone format"))
			},
			expectedUser:  nil,
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "invalid phone")
			},
		},
		{
			name: "user deleted",
			filter: &domain.UserFilter{
				ID: &userID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "username", "phone", "password_hash", "salt",
					"created_at", "updated_at", "deleted_at", "last_seen",
				}).AddRow(
					userID, username, phone, passwordHash, salt,
					now, now, now.Unix(), now,
				)

				mock.ExpectQuery("SELECT (.+) FROM users").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expectedUser: &domain.User{
				ID:           userID,
				Username:     &username,
				Phone:        &phone,
				PasswordHash: passwordHash,
				Salt:         &salt,
				CreatedAt:    now,
				UpdatedAt:    now,
				DeletedAt:    now.Unix(),
				LastSeen:     &now,
			},
			expectedError: false,
		},
		{
			name: "multiple rows returned (should handle first)",
			filter: &domain.UserFilter{
				ID: &userID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				userID2 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
				rows := pgxmock.NewRows([]string{
					"id", "username", "phone", "password_hash", "salt",
					"created_at", "updated_at", "deleted_at", "last_seen",
				}).AddRow(
					userID, username, phone, passwordHash, salt,
					now, now, 0, now,
				).AddRow(
					userID2, username+"2", phone+"2", passwordHash+"2", salt+"2",
					now, now, 0, now,
				)

				mock.ExpectQuery("SELECT (.+) FROM users").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expectedUser: &domain.User{
				ID:           userID,
				Username:     &username,
				Phone:        &phone,
				PasswordHash: passwordHash,
				Salt:         &salt,
				CreatedAt:    now,
				UpdatedAt:    now,
				DeletedAt:    0,
				LastSeen:     &now,
			},
			expectedError: false,
		},
		{
			name: "null values in database",
			filter: &domain.UserFilter{
				Phone: &phone,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "username", "phone", "password_hash", "salt",
					"created_at", "updated_at", "deleted_at", "last_seen",
				}).AddRow(
					userID, nil, phone, passwordHash, nil,
					now, now, 0, nil,
				)

				mock.ExpectQuery("SELECT (.+) FROM users").
					WithArgs(phone).
					WillReturnRows(rows)
			},
			expectedUser: &domain.User{
				ID:           userID,
				Username:     nil,
				Phone:        &phone,
				PasswordHash: passwordHash,
				Salt:         nil,
				CreatedAt:    now,
				UpdatedAt:    now,
				DeletedAt:    0,
				LastSeen:     nil,
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock pool
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			// Setup mock expectations
			tt.setupMock(mockPool)

			// Create repository
			repo := &Repo{
				pool:    mockPool,
				builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
				logger:  logger.New("debug"),
			}

			// Execute test
			user, err := repo.Get(ctx, tt.filter)

			// Verify expectations
			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, user)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tt.expectedUser.ID, user.ID)
				assert.Equal(t, tt.expectedUser.Phone, user.Phone)
				assert.Equal(t, tt.expectedUser.PasswordHash, user.PasswordHash)
			}

			// Ensure all expectations were met
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestRepo_GetByPhone(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	phone := "+998901234567"
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	username := "testuser"
	passwordHash := "hashed_password"
	salt := "salt123"

	tests := []struct {
		name          string
		phone         string
		setupMock     func(pgxmock.PgxPoolIface)
		expectedUser  *domain.User
		expectedError bool
	}{
		{
			name:  "success",
			phone: phone,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "username", "phone", "password_hash", "salt",
					"created_at", "updated_at", "deleted_at", "last_seen",
				}).AddRow(
					userID, username, phone, passwordHash, salt,
					now, now, 0, now,
				)

				mock.ExpectQuery("SELECT (.+) FROM users").
					WithArgs(phone).
					WillReturnRows(rows)
			},
			expectedUser: &domain.User{
				ID:           userID,
				Username:     &username,
				Phone:        &phone,
				PasswordHash: passwordHash,
				Salt:         &salt,
				CreatedAt:    now,
				UpdatedAt:    now,
				DeletedAt:    0,
				LastSeen:     &now,
			},
			expectedError: false,
		},
		{
			name:  "user not found",
			phone: phone,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM users").
					WithArgs(phone).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedUser:  nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock pool
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			// Setup mock expectations
			tt.setupMock(mockPool)

			// Create repository
			repo := &Repo{
				pool:    mockPool,
				builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
				logger:  logger.New("debug"),
			}

			// Execute test
			user, err := repo.GetByPhone(ctx, tt.phone)

			// Verify expectations
			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tt.expectedUser.ID, user.ID)
				assert.Equal(t, tt.expectedUser.Phone, user.Phone)
			}

			// Ensure all expectations were met
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
