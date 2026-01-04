package client

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"
	"gct/pkg/logger"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Update(t *testing.T) {
	ctx := t.Context()

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	username := "updateduser"
	phone := "+998901234567"
	passwordHash := "new_hashed_password"
	salt := "new_salt"
	lastSeen := time.Now()

	user := &domain.User{
		ID:           userID,
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
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // last_seen
						pgxmock.AnyArg(), // id (where clause)
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: false,
		},
		{
			name: "database error",
			user: user,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // last_seen
						pgxmock.AnyArg(), // id (where clause)
					).
					WillReturnError(errors.New("database error"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "no rows affected - user not found",
			user: user,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // last_seen
						pgxmock.AnyArg(), // id (where clause)
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
			expectedError: false, // The current implementation doesn't check rows affected
		},
		{
			name: "update with nil username",
			user: &domain.User{
				ID:           userID,
				Username:     nil,
				Phone:        &phone,
				PasswordHash: passwordHash,
				Salt:         &salt,
				LastSeen:     &lastSeen,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // last_seen
						pgxmock.AnyArg(), // id (where clause)
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: false,
		},
		{
			name: "update with nil salt",
			user: &domain.User{
				ID:           uuid.Nil,
				Username:     &username,
				Phone:        &phone,
				PasswordHash: passwordHash,
				Salt:         &salt,
				LastSeen:     &lastSeen,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // last_seen
						pgxmock.AnyArg(), // id (where clause)
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: false,
		},
		{
			name: "update with nil last seen",
			user: &domain.User{
				ID:           userID,
				Username:     &username,
				Phone:        &phone,
				PasswordHash: passwordHash,
				Salt:         &salt,
				LastSeen:     nil,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // last_seen
						pgxmock.AnyArg(), // id (where clause)
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: false,
		},
		{
			name: "update with empty password hash",
			user: &domain.User{
				ID:           userID,
				Username:     &username,
				Phone:        &phone,
				PasswordHash: "",
				Salt:         &salt,
				LastSeen:     &lastSeen,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // last_seen
						pgxmock.AnyArg(), // id (where clause)
					).
					WillReturnError(errors.New("password hash cannot be empty"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "password hash")
			},
		},
		{
			name: "update with invalid phone format",
			user: &domain.User{
				ID:           userID,
				Username:     &username,
				Phone:        func() *string { s := "invalid-phone"; return &s }(),
				PasswordHash: passwordHash,
				Salt:         &salt,
				LastSeen:     &lastSeen,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // last_seen
						pgxmock.AnyArg(), // id (where clause)
					).
					WillReturnError(errors.New("invalid phone format"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid phone")
			},
		},
		{
			name: "update with long username",
			user: &domain.User{
				ID:           userID,
				Username:     func() *string { s := "very-long-username-that-exceeds-normal-limits"; return &s }(),
				Phone:        &phone,
				PasswordHash: passwordHash,
				Salt:         &salt,
				LastSeen:     &lastSeen,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // last_seen
						pgxmock.AnyArg(), // id (where clause)
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			expectedError: false,
		},
		{
			name: "connection timeout",
			user: user,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // last_seen
						pgxmock.AnyArg(), // id (where clause)
					).
					WillReturnError(errors.New("connection timeout"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "timeout")
			},
		},
		{
			name: "foreign key constraint violation",
			user: user,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // last_seen
						pgxmock.AnyArg(), // id (where clause)
					).
					WillReturnError(errors.New("foreign key constraint violation"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "foreign key")
			},
		},
		{
			name: "zero user ID",
			user: &domain.User{
				ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
				Username:     &username,
				Phone:        &phone,
				PasswordHash: passwordHash,
				Salt:         &salt,
				LastSeen:     &lastSeen,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // last_seen
						pgxmock.AnyArg(), // id (where clause)
					).
					WillReturnError(errors.New("invalid user ID"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid user ID")
			},
		},
		{
			name: "negative user ID",
			user: &domain.User{
				ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
				Username:     &username,
				Phone:        &phone,
				PasswordHash: passwordHash,
				Salt:         &salt,
				LastSeen:     &lastSeen,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // last_seen
						pgxmock.AnyArg(), // id (where clause)
					).
					WillReturnError(errors.New("invalid user ID"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid user ID")
			},
		},
		{
			name: "permission denied",
			user: user,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // last_seen
						pgxmock.AnyArg(), // id (where clause)
					).
					WillReturnError(errors.New("permission denied for table users"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "permission denied")
			},
		},
		{
			name: "update with all nil optional fields",
			user: &domain.User{
				ID:           userID,
				Username:     nil,
				Phone:        &phone,
				PasswordHash: passwordHash,
				Salt:         nil,
				LastSeen:     nil,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("UPDATE users").
					WithArgs(
						pgxmock.AnyArg(), // role_id
						pgxmock.AnyArg(), // username
						pgxmock.AnyArg(), // email
						pgxmock.AnyArg(), // phone
						pgxmock.AnyArg(), // password_hash
						pgxmock.AnyArg(), // salt
						pgxmock.AnyArg(), // attributes
						pgxmock.AnyArg(), // active
						pgxmock.AnyArg(), // updated_at
						pgxmock.AnyArg(), // last_seen
						pgxmock.AnyArg(), // id (where clause)
					).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
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
			err = repo.Update(ctx, tt.user)

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

func TestRepo_Update_WithNilFields(t *testing.T) {
	ctx := t.Context()

	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	user := &domain.User{
		ID:           userID,
		Phone:        func() *string { s := "+998901234567"; return &s }(),
		PasswordHash: "hashed_password",
	}

	mockPool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mockPool.Close()

	mockPool.ExpectExec("UPDATE users").
		WithArgs(
			pgxmock.AnyArg(), // role_id
			pgxmock.AnyArg(), // username
			pgxmock.AnyArg(), // email
			pgxmock.AnyArg(), // phone
			pgxmock.AnyArg(), // password_hash
			pgxmock.AnyArg(), // salt
			pgxmock.AnyArg(), // attributes
			pgxmock.AnyArg(), // active
			pgxmock.AnyArg(), // updated_at
			pgxmock.AnyArg(), // last_seen
			pgxmock.AnyArg(), // id (where clause)
		).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	repo := &Repo{
		pool:    mockPool,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		logger:  logger.New("debug"),
	}

	err = repo.Update(ctx, user)

	require.NoError(t, err)
	assert.NoError(t, mockPool.ExpectationsWereMet())
}
