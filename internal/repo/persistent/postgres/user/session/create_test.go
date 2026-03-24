package session

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
	ctx := t.Context()

	sessionID := uuid.New()
	deviceID := uuid.New()
	deviceName := "iPhone 14 Pro"
	deviceType := domain.DeviceTypeMobile
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0"
	fcmToken := "fcm_token_123"
	userID := uuid.New()
	// data := map[string]any{"ip": "192.168.1.1"}
	// dataJSON, _ := json.Marshal(data)

	sess := &domain.Session{
		ID:         sessionID,
		DeviceID:   deviceID,
		DeviceName: &deviceName,
		DeviceType: &deviceType,
		IPAddress:  &ipAddress,
		UserAgent:  &userAgent,
		FCMToken:   &fcmToken,
		UserID:     userID,
		// // Data:         dataJSON,
		Revoked:      false,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		LastActivity: time.Now(),
	}

	tests := []struct {
		name          string
		session       *domain.Session
		setupMock     func(pgxmock.PgxPoolIface)
		expectedError bool
		errorCheck    func(*testing.T, error)
	}{
		{
			name:    "success with provided ID",
			session: sess,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO session").
					WithArgs(
						pgxmock.AnyArg(), // id
						pgxmock.AnyArg(), // device_id
						pgxmock.AnyArg(), // device_name
						pgxmock.AnyArg(), // device_type
						pgxmock.AnyArg(), // ip_address
						pgxmock.AnyArg(), // user_agent
						pgxmock.AnyArg(), // fcm_token
						pgxmock.AnyArg(), // refresh_token_hash
						pgxmock.AnyArg(), // data
						pgxmock.AnyArg(), // user_id
						pgxmock.AnyArg(), // revoked
						pgxmock.AnyArg(), // expires_at
						pgxmock.AnyArg(), // last_activity
						pgxmock.AnyArg(), // created_at
						pgxmock.AnyArg(), // updated_at
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: false,
		},
		{
			name: "success - auto-generate ID",
			session: &domain.Session{
				ID:       uuid.Nil, // Will be auto-generated
				DeviceID: deviceID,
				UserID:   userID,
				// // Data:         dataJSON,
				ExpiresAt:    time.Now().Add(24 * time.Hour),
				LastActivity: time.Now(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: false,
		},
		{
			name:    "database error",
			session: sess,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnError(errors.New("database error"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "nil optional fields",
			session: &domain.Session{
				ID:         sessionID,
				DeviceID:   deviceID,
				DeviceName: nil,
				DeviceType: nil,
				IPAddress:  nil,
				UserAgent:  nil,
				FCMToken:   nil,
				UserID:     userID,
				// Data:         dataJSON,
				Revoked:      false,
				ExpiresAt:    time.Now().Add(24 * time.Hour),
				LastActivity: time.Now(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: false,
		},
		{
			name: "empty device name",
			session: &domain.Session{
				ID:         sessionID,
				DeviceID:   deviceID,
				DeviceName: func() *string { s := ""; return &s }(),
				DeviceType: &deviceType,
				IPAddress:  &ipAddress,
				UserAgent:  &userAgent,
				FCMToken:   &fcmToken,
				UserID:     userID,
				// Data:         dataJSON,
				Revoked:      false,
				ExpiresAt:    time.Now().Add(24 * time.Hour),
				LastActivity: time.Now(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: false,
		},
		{
			name: "zero user ID",
			session: &domain.Session{
				ID:         sessionID,
				DeviceID:   deviceID,
				DeviceName: &deviceName,
				DeviceType: &deviceType,
				IPAddress:  &ipAddress,
				UserAgent:  &userAgent,
				FCMToken:   &fcmToken,
				UserID:     uuid.Nil,
				// Data:         dataJSON,
				Revoked:      false,
				ExpiresAt:    time.Now().Add(24 * time.Hour),
				LastActivity: time.Now(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
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
			name: "negative company ID",
			session: &domain.Session{
				ID:         sessionID,
				DeviceID:   deviceID,
				DeviceName: &deviceName,
				DeviceType: &deviceType,
				IPAddress:  &ipAddress,
				UserAgent:  &userAgent,
				FCMToken:   &fcmToken,
				UserID:     uuid.Nil,
				// Data:         dataJSON,
				Revoked:      false,
				ExpiresAt:    time.Now().Add(24 * time.Hour),
				LastActivity: time.Now(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnError(errors.New("invalid company ID"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid company ID")
			},
		},
		{
			name: "expired session time",
			session: &domain.Session{
				ID:         sessionID,
				DeviceID:   deviceID,
				DeviceName: &deviceName,
				DeviceType: &deviceType,
				IPAddress:  &ipAddress,
				UserAgent:  &userAgent,
				FCMToken:   &fcmToken,
				UserID:     userID,
				// // Data:         dataJSON,
				Revoked:      false,
				ExpiresAt:    time.Now().Add(-1 * time.Hour), // Already expired
				LastActivity: time.Now(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: false,
		},
		{
			name:    "connection timeout",
			session: sess,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
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
			name:    "foreign key constraint violation",
			session: sess,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
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
			name:    "unique constraint violation",
			session: sess,
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnError(errors.New("unique constraint violation"))
			},
			expectedError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unique constraint")
			},
		},
		{
			name: "desktop device type",
			session: &domain.Session{
				ID:         sessionID,
				DeviceID:   deviceID,
				DeviceName: &deviceName,
				DeviceType: func() *domain.SessionDeviceType { dt := domain.DeviceTypeDesktop; return &dt }(),
				IPAddress:  &ipAddress,
				UserAgent:  &userAgent,
				FCMToken:   &fcmToken,
				UserID:     userID,
				// Data:         dataJSON,
				Revoked:      false,
				ExpiresAt:    time.Now().Add(24 * time.Hour),
				LastActivity: time.Now(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
			expectedError: false,
		},
		{
			name: "revoked session creation",
			session: &domain.Session{
				ID:         sessionID,
				DeviceID:   deviceID,
				DeviceName: &deviceName,
				DeviceType: &deviceType,
				IPAddress:  &ipAddress,
				UserAgent:  &userAgent,
				FCMToken:   &fcmToken,
				UserID:     userID,
				// // Data:         dataJSON,
				Revoked:      true, // Created as revoked
				ExpiresAt:    time.Now().Add(24 * time.Hour),
				LastActivity: time.Now(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectExec("INSERT INTO session").
					WithArgs(
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
						pgxmock.AnyArg(),
					).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
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
			err = repo.Create(ctx, tt.session)

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
