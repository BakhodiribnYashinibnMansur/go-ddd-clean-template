package client

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gct/internal/domain"
	"gct/pkg/logger"
)

func TestRepo_Create(t *testing.T) {
	ctx := context.Background()
	lastSeen := time.Now().Add(-1 * time.Hour)

	username := "testuser"
	phone := "+998901234567"
	passwordHash := "hashed_password"
	salt := "salt123"

	user := &domain.User{
		Username:     &username,
		Phone:        &phone,
		PasswordHash: passwordHash,
		Salt:         &salt,
		LastSeen:     &lastSeen,
	}

	tests := []struct {
		name          string
		user          *domain.User
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name: "success",
			user: user,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // created_at
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // deleted_at
						pgxmock.AnyArg(), // last_seen
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: false,
		},
		{
			name: "database error",
			user: user,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // created_at
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // deleted_at
						pgxmock.AnyArg(), // last_seen
					).
					WillReturnError(errors.New("database error"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.NotNil(t, err)
			},
		},
		{
			name: "no rows affected",
			user: user,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // created_at
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // deleted_at
						pgxmock.AnyArg(), // last_seen
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 0))
			},
			expectedError: false,
		},
		{
			name: "duplicate phone error",
			user: user,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // created_at
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // deleted_at
						pgxmock.AnyArg(), // last_seen
					).
					WillReturnError(errors.New("duplicate key value violates unique constraint"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "duplicate key")
			},
		},
		{
			name: "invalid phone format",
			user: &domain.User{
				Username:     &username,
				Phone:        func() *string { s := "invalid-phone"; return &s }(),
				PasswordHash: passwordHash,
				Salt:         &salt,
				LastSeen:     &lastSeen,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // created_at
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // deleted_at
						pgxmock.AnyArg(), // last_seen
					).
					WillReturnError(errors.New("invalid phone format"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "invalid phone")
			},
		},
		{
			name: "empty password hash",
			user: &domain.User{
				Username:     &username,
				Phone:        &phone,
				PasswordHash: "",
				Salt:         &salt,
				LastSeen:     &lastSeen,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // created_at
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // deleted_at
						pgxmock.AnyArg(), // last_seen
					).
					WillReturnError(errors.New("password hash cannot be empty"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "password hash")
			},
		},
		{
			name: "nil username",
			user: &domain.User{
				Username:     nil,
				Phone:        &phone,
				PasswordHash: passwordHash,
				Salt:         &salt,
				LastSeen:     &lastSeen,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // created_at
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // deleted_at
						pgxmock.AnyArg(), // last_seen
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: false,
		},
		{
			name: "nil salt",
			user: &domain.User{
				Username:     &username,
				Phone:        &phone,
				PasswordHash: passwordHash,
				Salt:         nil,
				LastSeen:     &lastSeen,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // created_at
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // deleted_at
						pgxmock.AnyArg(), // last_seen
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: false,
		},
		{
			name: "nil last seen",
			user: &domain.User{
				Username:     &username,
				Phone:        &phone,
				PasswordHash: passwordHash,
				Salt:         &salt,
				LastSeen:     nil,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // created_at
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // deleted_at
						pgxmock.AnyArg(), // last_seen
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: false,
		},
		{
			name: "long username",
			user: &domain.User{
				Username:     func() *string { s := strings.Repeat("a", 100); return &s }(),
				Phone:        &phone,
				PasswordHash: passwordHash,
				Salt:         &salt,
				LastSeen:     &lastSeen,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // created_at
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // deleted_at
						pgxmock.AnyArg(), // last_seen
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: false,
		},
		{
			name: "connection timeout",
			user: user,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // created_at
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // deleted_at
						pgxmock.AnyArg(), // last_seen
					).
					WillReturnError(errors.New("connection timeout"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "timeout")
			},
		},
		{
			name: "foreign key constraint violation",
			user: user,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // created_at
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // deleted_at
						pgxmock.AnyArg(), // last_seen
					).
					WillReturnError(errors.New("foreign key constraint violation"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "foreign key")
			},
		},
		{
			name: "check constraint violation",
			user: &domain.User{
				Username:     &username,
				Phone:        func() *string { s := "+998901234567"; return &s }(),
				PasswordHash: passwordHash,
				Salt:         &salt,
				LastSeen:     &lastSeen,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // created_at
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // deleted_at
						pgxmock.AnyArg(), // last_seen
					).
					WillReturnError(errors.New("check constraint violation"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), "check constraint")
			},
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

			// Create repository with mock pool
			repo := &Repo{
				pool:    mockPool,
				builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
				logger:  logger.New("debug"),
			}

			// Execute test
			err = repo.Create(ctx, tt.user)

			// Verify expectations
			if tt.expectedError {
				require.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
			}

			// Ensure all expectations were met
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestRepo_Create_WithNilFields(t *testing.T) {
	ctx := context.Background()

	phone := "+998901234567"
	user := &domain.User{
		Phone:        &phone,
		PasswordHash: "hashed_password",
	}

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	mockPool.ExpectExec("INSERT INTO users").
		WithArgs(
			pgxmock.AnyArg(), // id
			pgxmock.AnyArg(), // role_id
			pgxmock.AnyArg(), // username
			pgxmock.AnyArg(), // email
			pgxmock.AnyArg(), // phone
			pgxmock.AnyArg(), // password_hash
			pgxmock.AnyArg(), // salt
			pgxmock.AnyArg(), // attributes
			pgxmock.AnyArg(), // active
			pgxmock.AnyArg(), // created_at
			pgxmock.AnyArg(), // updated_at
			pgxmock.AnyArg(), // deleted_at
			pgxmock.AnyArg(), // last_seen
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	err = repo.Create(ctx, user)

	require.NoError(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
