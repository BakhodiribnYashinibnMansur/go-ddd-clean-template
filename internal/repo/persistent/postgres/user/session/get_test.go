package session

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"
	"gct/pkg/logger"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepo_Get(t *testing.T) {
	ctx := t.Context()
	now := time.Now()

	sID := uuid.New()
	deviceID := uuid.New()
	deviceName := "Test Device"
	deviceType := domain.DeviceTypeMobile
	ipAddress := "127.0.0.1"
	userAgent := "TestAgent"
	fcmToken := "token"

	tests := []struct {
		name            string
		filter          *domain.SessionFilter
		setupMock       func(pgxmock.PgxPoolIface)
		expectedSession *domain.Session
		expectedError   bool
	}{
		{
			name: "success",
			filter: &domain.SessionFilter{
				ID: &sID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "device_id", "device_name", "device_type", "ip_address", "user_agent",
					"fcm_token", "refresh_token_hash", "data", "user_id", "revoked", "expires_at", "last_activity", "created_at", "updated_at",
				}).AddRow(
					sID, deviceID, nil, nil, nil, nil,
					nil, "refresh_hash", []byte("{}"), uuid.New(), false, now, now, now, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM session").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedSession: &domain.Session{
				ID:               sID,
				DeviceID:         deviceID,
				DeviceName:       &deviceName,
				DeviceType:       &deviceType,
				IPAddress:        &ipAddress,
				UserAgent:        &userAgent,
				FCMToken:         &fcmToken,
				RefreshTokenHash: "refresh_hash",
				Data:             []byte("{}"),
				UserID:           uuid.New(),
				Revoked:          false,
				ExpiresAt:        now,
				LastActivity:     now,
				CreatedAt:        now,
				UpdatedAt:        now,
			},
			expectedError: false,
		},
		{
			name: "not found",
			filter: &domain.SessionFilter{
				ID: &sID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM session").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(pgx.ErrNoRows)
			},
			expectedSession: nil,
			expectedError:   true,
		},
		{
			name: "db error",
			filter: &domain.SessionFilter{
				ID: &sID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM session").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("db error"))
			},
			expectedSession: nil,
			expectedError:   true,
		},
		{
			name: "get by user ID",
			filter: &domain.SessionFilter{
				UserID: func() *uuid.UUID { id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"); return &id }(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "device_id", "device_name", "device_type", "ip_address", "user_agent",
					"fcm_token", "refresh_token_hash", "data", "user_id", "revoked", "expires_at", "last_activity", "created_at", "updated_at",
				}).AddRow(
					sID, deviceID, nil, nil, nil, nil,
					nil, "refresh_hash", []byte("{}"), uuid.New(), false, now, now, now, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM session").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedSession: &domain.Session{
				ID:           sID,
				DeviceID:     deviceID,
				DeviceName:   &deviceName,
				DeviceType:   &deviceType,
				IPAddress:    &ipAddress,
				UserAgent:    &userAgent,
				FCMToken:     &fcmToken,
				ExpiresAt:    now,
				LastActivity: now,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			expectedError: false,
		},
		{
			name: "get revoked session",
			filter: &domain.SessionFilter{
				ID: &sID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "device_id", "device_name", "device_type", "ip_address", "user_agent",
					"fcm_token", "refresh_token_hash", "data", "user_id", "revoked", "expires_at", "last_activity", "created_at", "updated_at",
				}).AddRow(
					sID, deviceID, nil, nil, nil, nil,
					nil, "refresh_hash", []byte("{}"), uuid.New(), false, now, now, now, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM session").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedSession: &domain.Session{
				ID:           sID,
				DeviceID:     deviceID,
				DeviceName:   &deviceName,
				DeviceType:   &deviceType,
				IPAddress:    &ipAddress,
				UserAgent:    &userAgent,
				FCMToken:     &fcmToken,
				ExpiresAt:    now,
				LastActivity: now,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			expectedError: false,
		},
		{
			name: "session with nil optional fields",
			filter: &domain.SessionFilter{
				ID: &sID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "device_id", "device_name", "device_type", "ip_address", "user_agent",
					"fcm_token", "refresh_token_hash", "data", "user_id", "revoked", "expires_at", "last_activity", "created_at", "updated_at",
				}).AddRow(
					sID, deviceID, nil, nil, nil, nil,
					nil, "refresh_hash", []byte("{}"), uuid.New(), false, now, now, now, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM session").
					WithArgs(pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedSession: &domain.Session{
				ID:           sID,
				DeviceID:     deviceID,
				DeviceName:   nil,
				DeviceType:   nil,
				IPAddress:    nil,
				UserAgent:    nil,
				FCMToken:     nil,
				ExpiresAt:    now,
				LastActivity: now,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			expectedError: false,
		},
		{
			name: "connection timeout",
			filter: &domain.SessionFilter{
				ID: &sID,
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT (.+) FROM session").
					WithArgs(pgxmock.AnyArg()).
					WillReturnError(errors.New("connection timeout"))
			},
			expectedSession: nil,
			expectedError:   true,
		},
		{
			name: "get by revoked filter",
			filter: &domain.SessionFilter{
				Revoked: func() *bool { r := true; return &r }(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "device_id", "device_name", "device_type", "ip_address", "user_agent",
					"fcm_token", "refresh_token_hash", "data", "user_id", "revoked", "expires_at", "last_activity", "created_at", "updated_at",
				}).AddRow(
					sID, deviceID, deviceName, deviceType, ipAddress, userAgent,
					fcmToken, "refresh_hash", []byte("{}"), uuid.New(), false, now, now, now, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM session").
					WithArgs(true).
					WillReturnRows(rows)
			},
			expectedSession: &domain.Session{
				ID:           sID,
				DeviceID:     deviceID,
				DeviceName:   &deviceName,
				DeviceType:   &deviceType,
				IPAddress:    &ipAddress,
				UserAgent:    &userAgent,
				FCMToken:     &fcmToken,
				ExpiresAt:    now,
				LastActivity: now,
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			expectedError: false,
		},
		{
			name:   "empty filter",
			filter: &domain.SessionFilter{},
			setupMock: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{
					"id", "device_id", "device_name", "device_type", "ip_address", "user_agent",
					"fcm_token", "refresh_token_hash", "data", "user_id", "revoked", "expires_at", "last_activity", "created_at", "updated_at",
				}).AddRow(
					sID, deviceID, nil, nil, nil, nil,
					nil, "refresh_hash", []byte("{}"), uuid.New(), false, now, now, now, now,
				)
				mock.ExpectQuery("SELECT (.+) FROM session").
					WillReturnRows(rows)
			},
			expectedSession: &domain.Session{
				ID:           sID,
				DeviceID:     deviceID,
				DeviceName:   &deviceName,
				DeviceType:   &deviceType,
				IPAddress:    &ipAddress,
				UserAgent:    &userAgent,
				FCMToken:     &fcmToken,
				ExpiresAt:    now,
				LastActivity: now,
				CreatedAt:    now,
				UpdatedAt:    now,
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

			s, err := repo.Get(ctx, tt.filter)

			if tt.expectedError {
				require.Error(t, err)
				require.Nil(t, s)
			} else {
				require.NoError(t, err)
				require.NotNil(t, s)
				assert.Equal(t, tt.expectedSession.ID, s.ID)
				assert.Equal(t, tt.expectedSession.DeviceName, s.DeviceName)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}
