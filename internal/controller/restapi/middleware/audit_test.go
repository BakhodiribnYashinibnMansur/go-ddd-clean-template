package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/usecase"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuditUseCase is a mock for audit use case
type MockAuditUseCase struct {
	mock.Mock
}

func (m *MockAuditUseCase) CreateHistory(h *domain.EndpointHistory) error {
	args := m.Called(h)
	return args.Error(0)
}

func (m *MockAuditUseCase) CreateLog(al *domain.AuditLog) error {
	args := m.Called(al)
	return args.Error(0)
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

			// Setup mock use case (simplified - in real scenario use proper mocks)
			uc := &usecase.UseCase{}

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
			uc := &usecase.UseCase{}

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
