package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/usecase"
	"gct/internal/usecase/audit"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEndpointHistoryUC is a mock for endpoint history use case
type MockEndpointHistoryUC struct {
	mock.Mock
}

func (m *MockEndpointHistoryUC) Create(ctx context.Context, h *domain.EndpointHistory) error {
	args := m.Called(ctx, h)
	return args.Error(0)
}

func (m *MockEndpointHistoryUC) Gets(ctx context.Context, in *domain.EndpointHistoriesFilter) ([]*domain.EndpointHistory, int, error) {
	args := m.Called(ctx, in)
	return args.Get(0).([]*domain.EndpointHistory), args.Int(1), args.Error(2)
}

type MockAuditLogUC struct {
	mock.Mock
}

func (m *MockAuditLogUC) Create(ctx context.Context, al *domain.AuditLog) error {
	args := m.Called(ctx, al)
	return args.Error(0)
}

func (m *MockAuditLogUC) Gets(ctx context.Context, filter *domain.AuditLogsFilter) ([]*domain.AuditLog, int, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*domain.AuditLog), args.Int(1), args.Error(2)
}

func TestAuditMiddleware_EndpointHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		path           string
		setupSession   bool
		sessionID      uuid.UUID
		userID         uuid.UUID
		expectedStatus int
		shouldRecord   bool
	}{
		{
			name:           "record_get_request_without_session",
			method:         "GET",
			path:           "/api/v1/users",
			setupSession:   false,
			expectedStatus: http.StatusOK,
			shouldRecord:   true,
		},
		{
			name:           "record_post_request_with_session",
			method:         "POST",
			path:           "/api/v1/users",
			setupSession:   true,
			sessionID:      uuid.New(),
			userID:         uuid.New(),
			expectedStatus: http.StatusCreated,
			shouldRecord:   true,
		},
		{
			name:           "record_delete_request_with_session",
			method:         "DELETE",
			path:           "/api/v1/users/123",
			setupSession:   true,
			sessionID:      uuid.New(),
			userID:         uuid.New(),
			expectedStatus: http.StatusNoContent,
			shouldRecord:   true,
		},
		{
			name:           "record_failed_request",
			method:         "POST",
			path:           "/api/v1/users",
			setupSession:   true,
			sessionID:      uuid.New(),
			userID:         uuid.New(),
			expectedStatus: http.StatusBadRequest,
			shouldRecord:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup logger
			mockLogger := logger.New("debug")

			// Setup mock use case
			mockHistoryUC := new(MockEndpointHistoryUC)
			if tt.shouldRecord {
				mockHistoryUC.On("Create", mock.Anything, mock.Anything).Return(nil)
			}

			// Initialize UseCase struct
			uc := &usecase.UseCase{
				Audit: &audit.UseCase{
					History: mockHistoryUC,
				},
			}

			// Create middleware
			auditM := NewAuditMiddleware(uc, mockLogger)

			// Setup router
			r := gin.New()
			r.Use(auditM.EndpointHistory())

			// Setup test handler
			r.Handle(tt.method, tt.path, func(c *gin.Context) {
				// Setup session if needed
				if tt.setupSession {
					session := &domain.Session{
						ID:     tt.sessionID,
						UserID: tt.userID,
					}
					c.Set(consts.CtxSession, session)
				}

				// Return the expected status
				c.Status(tt.expectedStatus)
			})

			// Create request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("X-Request-ID", uuid.New().String())
			req.Header.Set("User-Agent", "test-agent")

			// Execute request
			r.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Give goroutine time to execute (in real tests, use proper synchronization)
			time.Sleep(100 * time.Millisecond)
		})
	}
}

func TestAuditMiddleware_ChangeAudit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		path           string
		setupSession   bool
		sessionID      uuid.UUID
		userID         uuid.UUID
		expectedStatus int
		shouldAudit    bool
	}{
		{
			name:           "skip_get_request",
			method:         "GET",
			path:           "/api/v1/users",
			setupSession:   false,
			expectedStatus: http.StatusOK,
			shouldAudit:    false,
		},
		{
			name:           "skip_head_request",
			method:         "HEAD",
			path:           "/api/v1/users",
			setupSession:   false,
			expectedStatus: http.StatusOK,
			shouldAudit:    false,
		},
		{
			name:           "skip_options_request",
			method:         "OPTIONS",
			path:           "/api/v1/users",
			setupSession:   false,
			expectedStatus: http.StatusOK,
			shouldAudit:    false,
		},
		{
			name:           "audit_post_request",
			method:         "POST",
			path:           "/api/v1/users",
			setupSession:   true,
			sessionID:      uuid.New(),
			userID:         uuid.New(),
			expectedStatus: http.StatusCreated,
			shouldAudit:    true,
		},
		{
			name:           "audit_put_request",
			method:         "PUT",
			path:           "/api/v1/users/123",
			setupSession:   true,
			sessionID:      uuid.New(),
			userID:         uuid.New(),
			expectedStatus: http.StatusOK,
			shouldAudit:    true,
		},
		{
			name:           "audit_delete_request",
			method:         "DELETE",
			path:           "/api/v1/users/123",
			setupSession:   true,
			sessionID:      uuid.New(),
			userID:         uuid.New(),
			expectedStatus: http.StatusNoContent,
			shouldAudit:    true,
		},
		{
			name:           "audit_patch_request",
			method:         "PATCH",
			path:           "/api/v1/users/123",
			setupSession:   true,
			sessionID:      uuid.New(),
			userID:         uuid.New(),
			expectedStatus: http.StatusOK,
			shouldAudit:    true,
		},
		{
			name:           "audit_failed_mutation",
			method:         "POST",
			path:           "/api/v1/users",
			setupSession:   true,
			sessionID:      uuid.New(),
			userID:         uuid.New(),
			expectedStatus: http.StatusBadRequest,
			shouldAudit:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup logger
			mockLogger := logger.New("debug")

			// Setup mock use case
			mockLogUC := new(MockAuditLogUC)
			if tt.shouldAudit {
				mockLogUC.On("Create", mock.Anything, mock.Anything).Return(nil)
			}

			// Initialize UseCase struct
			uc := &usecase.UseCase{
				Audit: &audit.UseCase{
					Log: mockLogUC,
				},
			}

			// Create middleware
			auditM := NewAuditMiddleware(uc, mockLogger)

			// Setup router
			r := gin.New()
			r.Use(auditM.ChangeAudit())

			// Setup test handler
			r.Handle(tt.method, tt.path, func(c *gin.Context) {
				// Setup session if needed
				if tt.setupSession {
					session := &domain.Session{
						ID:     tt.sessionID,
						UserID: tt.userID,
					}
					c.Set(consts.CtxSession, session)
				}

				// Return the expected status
				c.Status(tt.expectedStatus)
			})

			// Create request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path+"?test=param", nil)
			req.Header.Set("User-Agent", "test-agent")

			// Execute request
			r.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Give goroutine time to execute
			time.Sleep(100 * time.Millisecond)
		})
	}
}
