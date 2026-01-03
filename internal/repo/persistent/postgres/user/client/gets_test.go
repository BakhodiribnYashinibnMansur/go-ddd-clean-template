package client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gct/internal/domain"
	"gct/pkg/logger"
)

func TestRepo_Gets(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	phone := "+998901234567"
	username := "testuser"
	passwordHash := "hashed_password"
	salt := "salt123"

	tests := []struct {
		name          string
		filter        *domain.UsersFilter
		setupMock     func(pgxmock.PgxPoolIface)
		expectedUsers []*domain.User
		expectedCount int
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success - get all users",
			filter: &domain.UsersFilter{
				UserFilter: domain.UserFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				// Mock select query first (implementation runs this first)
				rows := pgxmock.NewRows([]string{
					"id", "role_id", "username", "email", "phone", "password_hash", "salt", "attributes", "active", "created_at", "updated_at", "deleted_at", "last_seen",
				}).AddRow(
					userID, nil, username, nil, phone, passwordHash, salt, map[string]any{}, true,
					now, now, 0, &now,
				).AddRow(
					uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"), nil, username+"2", nil, phone+"2", passwordHash+"2", salt+"2", map[string]any{}, true,
					now, now, 0, &now,
				)

				mock.ExpectQuery("SELECT id, role_id, username, email, phone, password_hash, salt, attributes, active, created_at, updated_at, deleted_at, last_seen FROM users WHERE deleted_at = 0").
					WillReturnRows(rows)

				// Mock count query second
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users WHERE deleted_at = 0").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(2))
			},
			expectedUsers: []*domain.User{
				{
					ID:           userID,
					RoleID:       nil,
					Username:     &username,
					Email:        nil,
					Phone:        &phone,
					PasswordHash: passwordHash,
					Salt:         &salt,
					Attributes:   map[string]any{},
					Active:       true,
					LastSeen:     &now,
					DeletedAt:    0,
					CreatedAt:    now,
					UpdatedAt:    now,
				},
				{
					ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
					RoleID:       nil,
					Username:     func() *string { s := username + "2"; return &s }(),
					Email:        nil,
					Phone:        func() *string { s := phone + "2"; return &s }(),
					PasswordHash: passwordHash + "2",
					Salt:         func() *string { s := salt + "2"; return &s }(),
					Attributes:   map[string]any{},
					Active:       true,
					LastSeen:     &now,
					DeletedAt:    0,
					CreatedAt:    now,
					UpdatedAt:    now,
				},
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "success - get by ID",
			filter: &domain.UsersFilter{
				UserFilter: domain.UserFilter{
					ID: func() *uuid.UUID { id := userID; return &id }(),
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				// Mock count query
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
					WithArgs(userID).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))

				// Mock select query
				rows := pgxmock.NewRows([]string{
					"id", "username", "phone", "password_hash", "salt",
					"created_at", "updated_at", "deleted_at", "last_seen",
				}).AddRow(
					userID, username, phone, passwordHash, salt,
					now, now, 0, &now,
				)

				mock.ExpectQuery("SELECT id, username, phone, password_hash, salt, created_at, updated_at, deleted_at, last_seen FROM users").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			expectedUsers: []*domain.User{
				{
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
			},
			expectedCount: 1,
			expectedError: false,
		},
		{
			name: "success - get by phone",
			filter: &domain.UsersFilter{
				UserFilter: domain.UserFilter{
					Phone: &phone,
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				// Mock count query
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
					WithArgs(phone).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))

				// Mock select query
				rows := pgxmock.NewRows([]string{
					"id", "username", "phone", "password_hash", "salt",
					"created_at", "updated_at", "deleted_at", "last_seen",
				}).AddRow(
					userID, username, phone, passwordHash, salt,
					now, now, 0, &now,
				)

				mock.ExpectQuery("SELECT id, username, phone, password_hash, salt, created_at, updated_at, deleted_at, last_seen FROM users").
					WithArgs(phone).
					WillReturnRows(rows)
			},
			expectedUsers: []*domain.User{
				{
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
			},
			expectedCount: 1,
			expectedError: false,
		},
		{
			name: "success - with pagination",
			filter: &domain.UsersFilter{
				UserFilter: domain.UserFilter{},
				Pagination: &domain.Pagination{
					Limit:  10,
					Offset: 0,
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				// Mock count query
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(25))

				// Mock select query
				rows := pgxmock.NewRows([]string{
					"id", "username", "phone", "password_hash", "salt",
					"created_at", "updated_at", "deleted_at", "last_seen",
				}).AddRow(
					userID, username, phone, passwordHash, salt,
					now, now, 0, &now,
				)

				mock.ExpectQuery("SELECT id, username, phone, password_hash, salt, created_at, updated_at, deleted_at, last_seen FROM users").
					WillReturnRows(rows)
			},
			expectedUsers: []*domain.User{
				{
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
			},
			expectedCount: 25,
			expectedError: false,
		},
		{
			name: "success - empty result",
			filter: &domain.UsersFilter{
				UserFilter: domain.UserFilter{
					ID: func() *uuid.UUID { id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440999"); return &id }(),
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				// Mock count query
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
					WithArgs(uuid.MustParse("550e8400-e29b-41d4-a716-446655440999")).
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))

				// Mock select query
				rows := pgxmock.NewRows([]string{
					"id", "username", "phone", "password_hash", "salt",
					"created_at", "updated_at", "deleted_at", "last_seen",
				})

				mock.ExpectQuery("SELECT id, username, phone, password_hash, salt, created_at, updated_at, deleted_at, last_seen FROM users").
					WithArgs(uuid.MustParse("550e8400-e29b-41d4-a716-446655440999")).
					WillReturnRows(rows)
			},
			expectedUsers: []*domain.User{},
			expectedCount: 0,
			expectedError: false,
		},
		{
			name: "success - users with nil fields",
			filter: &domain.UsersFilter{
				UserFilter: domain.UserFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				// Mock count query
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))

				// Mock select query
				rows := pgxmock.NewRows([]string{
					"id", "username", "phone", "password_hash", "salt",
					"created_at", "updated_at", "deleted_at", "last_seen",
				}).AddRow(
					userID, nil, phone, passwordHash, nil,
					now, now, 0, nil,
				)

				mock.ExpectQuery("SELECT id, username, phone, password_hash, salt, created_at, updated_at, deleted_at, last_seen FROM users").
					WillReturnRows(rows)
			},
			expectedUsers: []*domain.User{
				{
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
			},
			expectedCount: 1,
			expectedError: false,
		},
		{
			name: "database error on count",
			filter: &domain.UsersFilter{
				UserFilter: domain.UserFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				// Mock count query error
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
					WillReturnError(errors.New("database connection error"))
			},
			expectedUsers: nil,
			expectedCount: 0,
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.NotNil(t, err)
			},
		},
		{
			name: "database error on select",
			filter: &domain.UsersFilter{
				UserFilter: domain.UserFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				// Mock count query
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))

				// Mock select query error
				mock.ExpectQuery("SELECT id, username, phone, password_hash, salt, created_at, updated_at, deleted_at, last_seen FROM users").
					WillReturnError(errors.New("select query error"))
			},
			expectedUsers: nil,
			expectedCount: 0,
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.NotNil(t, err)
			},
		},
		{
			name: "connection timeout",
			filter: &domain.UsersFilter{
				UserFilter: domain.UserFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				// Mock count query timeout
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
					WillReturnError(errors.New("connection timeout"))
			},
			expectedUsers: nil,
			expectedCount: 0,
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "timeout")
			},
		},
		{
			name: "success - with limit and offset",
			filter: &domain.UsersFilter{
				UserFilter: domain.UserFilter{},
				Pagination: &domain.Pagination{
					Limit:  5,
					Offset: 10,
				},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				// Mock count query
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(50))

				// Mock select query
				rows := pgxmock.NewRows([]string{
					"id", "username", "phone", "password_hash", "salt",
					"created_at", "updated_at", "deleted_at", "last_seen",
				}).AddRow(
					userID, username, phone, passwordHash, salt,
					now, now, 0, &now,
				)

				mock.ExpectQuery("SELECT id, username, phone, password_hash, salt, created_at, updated_at, deleted_at, last_seen FROM users").
					WillReturnRows(rows)
			},
			expectedUsers: []*domain.User{
				{
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
			},
			expectedCount: 50,
			expectedError: false,
		},
		{
			name: "success - get deleted users (should return empty)",
			filter: &domain.UsersFilter{
				UserFilter: domain.UserFilter{},
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				// Mock count query - should return 0 since deleted_at = 0
				mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM users").
					WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(0))

				// Mock select query - should return empty since deleted_at = 0
				rows := pgxmock.NewRows([]string{
					"id", "username", "phone", "password_hash", "salt",
					"created_at", "updated_at", "deleted_at", "last_seen",
				})

				mock.ExpectQuery("SELECT id, username, phone, password_hash, salt, created_at, updated_at, deleted_at, last_seen FROM users").
					WillReturnRows(rows)
			},
			expectedUsers: []*domain.User{},
			expectedCount: 0,
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
			users, count, err := repo.Gets(ctx, tt.filter)

			// Verify expectations
			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, users)
				assert.Equal(t, 0, count)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
				require.Equal(t, len(tt.expectedUsers), len(users))

				for i, expectedUser := range tt.expectedUsers {
					assert.Equal(t, expectedUser.ID, users[i].ID)
					assert.Equal(t, expectedUser.Phone, users[i].Phone)
					assert.Equal(t, expectedUser.PasswordHash, users[i].PasswordHash)

					if expectedUser.Username != nil {
						require.NotNil(t, users[i].Username)
						assert.Equal(t, *expectedUser.Username, *users[i].Username)
					} else {
						assert.Nil(t, users[i].Username)
					}

					if expectedUser.Salt != nil {
						require.NotNil(t, users[i].Salt)
						assert.Equal(t, *expectedUser.Salt, *users[i].Salt)
					} else {
						assert.Nil(t, users[i].Salt)
					}

					if expectedUser.LastSeen != nil {
						require.NotNil(t, users[i].LastSeen)
						assert.Equal(t, *expectedUser.LastSeen, *users[i].LastSeen)
					} else {
						assert.Nil(t, users[i].LastSeen)
					}
				}
			}

			// Ensure all expectations were met
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
