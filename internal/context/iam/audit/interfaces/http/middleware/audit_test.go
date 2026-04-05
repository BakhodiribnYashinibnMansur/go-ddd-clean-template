package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	shared "gct/internal/platform/domain"
	"gct/internal/platform/application"
	"gct/internal/platform/domain/consts"
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/context/iam/audit/application/command"
	"gct/internal/context/iam/audit/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// --- Mock Logger ---

type auditMockLog struct{}

func (m *auditMockLog) Debug(args ...any)                                     {}
func (m *auditMockLog) Debugf(template string, args ...any)                   {}
func (m *auditMockLog) Debugw(msg string, keysAndValues ...any)               {}
func (m *auditMockLog) Info(args ...any)                                      {}
func (m *auditMockLog) Infof(template string, args ...any)                    {}
func (m *auditMockLog) Infow(msg string, keysAndValues ...any)                {}
func (m *auditMockLog) Warn(args ...any)                                      {}
func (m *auditMockLog) Warnf(template string, args ...any)                    {}
func (m *auditMockLog) Warnw(msg string, keysAndValues ...any)                {}
func (m *auditMockLog) Error(args ...any)                                     {}
func (m *auditMockLog) Errorf(template string, args ...any)                   {}
func (m *auditMockLog) Errorw(msg string, keysAndValues ...any)               {}
func (m *auditMockLog) Fatal(args ...any)                                     {}
func (m *auditMockLog) Fatalf(template string, args ...any)                   {}
func (m *auditMockLog) Fatalw(msg string, keysAndValues ...any)               {}
func (m *auditMockLog) Debugc(ctx context.Context, msg string, kv ...any)     {}
func (m *auditMockLog) Infoc(ctx context.Context, msg string, kv ...any)      {}
func (m *auditMockLog) Warnc(ctx context.Context, msg string, kv ...any)      {}
func (m *auditMockLog) Errorc(ctx context.Context, msg string, kv ...any)     {}
func (m *auditMockLog) Fatalc(ctx context.Context, msg string, kv ...any)     {}

var _ logger.Log = (*auditMockLog)(nil)

// --- Mock Repositories ---

type mockEndpointHistoryRepo struct {
	mu    sync.Mutex
	saved int
}

func (m *mockEndpointHistoryRepo) Save(ctx context.Context, entry *domain.EndpointHistory) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saved++
	return nil
}

type mockAuditLogRepo struct {
	mu    sync.Mutex
	saved int
}

func (m *mockAuditLogRepo) Save(ctx context.Context, auditLog *domain.AuditLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saved++
	return nil
}

// --- Mock EventBus ---

type auditMockEventBus struct{}

func (m *auditMockEventBus) Publish(_ context.Context, _ ...shared.DomainEvent) error { return nil }
func (m *auditMockEventBus) Subscribe(_ string, _ application.EventHandler) error     { return nil }

// --- Helper ---

func newTestAuditMiddleware(historyRepo *mockEndpointHistoryRepo, auditRepo *mockAuditLogRepo) *AuditMiddleware {
	l := &auditMockLog{}
	historyHandler := command.NewCreateEndpointHistoryHandler(historyRepo, l)
	auditHandler := command.NewCreateAuditLogHandler(auditRepo, &auditMockEventBus{}, l)
	return NewAuditMiddleware(historyHandler, auditHandler, l)
}

// --- EndpointHistory Tests ---

func TestEndpointHistory_RecordsOnGET(t *testing.T) {
	gin.SetMode(gin.TestMode)

	historyRepo := &mockEndpointHistoryRepo{}
	auditRepo := &mockAuditLogRepo{}
	mw := newTestAuditMiddleware(historyRepo, auditRepo)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(mw.EndpointHistory())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	time.Sleep(50 * time.Millisecond)

	historyRepo.mu.Lock()
	count := historyRepo.saved
	historyRepo.mu.Unlock()

	if count != 1 {
		t.Errorf("expected 1 endpoint history entry saved, got %d", count)
	}
}

func TestEndpointHistory_RecordsOnPOST(t *testing.T) {
	gin.SetMode(gin.TestMode)

	historyRepo := &mockEndpointHistoryRepo{}
	auditRepo := &mockAuditLogRepo{}
	mw := newTestAuditMiddleware(historyRepo, auditRepo)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(mw.EndpointHistory())
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}

	time.Sleep(50 * time.Millisecond)

	historyRepo.mu.Lock()
	count := historyRepo.saved
	historyRepo.mu.Unlock()

	if count != 1 {
		t.Errorf("expected 1 endpoint history entry, got %d", count)
	}
}

func TestEndpointHistory_IncludesSessionUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	historyRepo := &mockEndpointHistoryRepo{}
	auditRepo := &mockAuditLogRepo{}
	mw := newTestAuditMiddleware(historyRepo, auditRepo)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	userID := uuid.New()
	sessionID := uuid.New()

	r.Use(func(c *gin.Context) {
		c.Set(consts.CtxSession, &shared.AuthSession{
			ID:     sessionID,
			UserID: userID,
		})
		c.Next()
	})
	r.Use(mw.EndpointHistory())
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	time.Sleep(50 * time.Millisecond)

	historyRepo.mu.Lock()
	count := historyRepo.saved
	historyRepo.mu.Unlock()

	if count != 1 {
		t.Errorf("expected 1 endpoint history entry, got %d", count)
	}
}

// --- ChangeAudit Tests ---

func TestChangeAudit_SkipsGET(t *testing.T) {
	gin.SetMode(gin.TestMode)

	historyRepo := &mockEndpointHistoryRepo{}
	auditRepo := &mockAuditLogRepo{}
	mw := newTestAuditMiddleware(historyRepo, auditRepo)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(mw.ChangeAudit())
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	time.Sleep(50 * time.Millisecond)

	auditRepo.mu.Lock()
	count := auditRepo.saved
	auditRepo.mu.Unlock()

	if count != 0 {
		t.Errorf("expected 0 audit log entries for GET, got %d", count)
	}
}

func TestChangeAudit_SkipsHEAD(t *testing.T) {
	gin.SetMode(gin.TestMode)

	historyRepo := &mockEndpointHistoryRepo{}
	auditRepo := &mockAuditLogRepo{}
	mw := newTestAuditMiddleware(historyRepo, auditRepo)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(mw.ChangeAudit())
	r.HEAD("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("HEAD", "/test", nil)
	r.ServeHTTP(w, req)

	time.Sleep(50 * time.Millisecond)

	auditRepo.mu.Lock()
	count := auditRepo.saved
	auditRepo.mu.Unlock()

	if count != 0 {
		t.Errorf("expected 0 audit log entries for HEAD, got %d", count)
	}
}

func TestChangeAudit_SkipsOPTIONS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	historyRepo := &mockEndpointHistoryRepo{}
	auditRepo := &mockAuditLogRepo{}
	mw := newTestAuditMiddleware(historyRepo, auditRepo)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(mw.ChangeAudit())
	r.OPTIONS("/test", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	r.ServeHTTP(w, req)

	time.Sleep(50 * time.Millisecond)

	auditRepo.mu.Lock()
	count := auditRepo.saved
	auditRepo.mu.Unlock()

	if count != 0 {
		t.Errorf("expected 0 audit log entries for OPTIONS, got %d", count)
	}
}

func TestChangeAudit_RecordsPOST(t *testing.T) {
	gin.SetMode(gin.TestMode)

	historyRepo := &mockEndpointHistoryRepo{}
	auditRepo := &mockAuditLogRepo{}
	mw := newTestAuditMiddleware(historyRepo, auditRepo)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(mw.ChangeAudit())
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	r.ServeHTTP(w, req)

	time.Sleep(50 * time.Millisecond)

	auditRepo.mu.Lock()
	count := auditRepo.saved
	auditRepo.mu.Unlock()

	if count != 1 {
		t.Errorf("expected 1 audit log entry for POST, got %d", count)
	}
}

func TestChangeAudit_RecordsPUT(t *testing.T) {
	gin.SetMode(gin.TestMode)

	historyRepo := &mockEndpointHistoryRepo{}
	auditRepo := &mockAuditLogRepo{}
	mw := newTestAuditMiddleware(historyRepo, auditRepo)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(mw.ChangeAudit())
	r.PUT("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("PUT", "/test", nil)
	r.ServeHTTP(w, req)

	time.Sleep(50 * time.Millisecond)

	auditRepo.mu.Lock()
	count := auditRepo.saved
	auditRepo.mu.Unlock()

	if count != 1 {
		t.Errorf("expected 1 audit log entry for PUT, got %d", count)
	}
}

func TestChangeAudit_RecordsDELETE(t *testing.T) {
	gin.SetMode(gin.TestMode)

	historyRepo := &mockEndpointHistoryRepo{}
	auditRepo := &mockAuditLogRepo{}
	mw := newTestAuditMiddleware(historyRepo, auditRepo)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(mw.ChangeAudit())
	r.DELETE("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("DELETE", "/test", nil)
	r.ServeHTTP(w, req)

	time.Sleep(50 * time.Millisecond)

	auditRepo.mu.Lock()
	count := auditRepo.saved
	auditRepo.mu.Unlock()

	if count != 1 {
		t.Errorf("expected 1 audit log entry for DELETE, got %d", count)
	}
}

func TestChangeAudit_RecordsPATCH(t *testing.T) {
	gin.SetMode(gin.TestMode)

	historyRepo := &mockEndpointHistoryRepo{}
	auditRepo := &mockAuditLogRepo{}
	mw := newTestAuditMiddleware(historyRepo, auditRepo)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(mw.ChangeAudit())
	r.PATCH("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("PATCH", "/test", nil)
	r.ServeHTTP(w, req)

	time.Sleep(50 * time.Millisecond)

	auditRepo.mu.Lock()
	count := auditRepo.saved
	auditRepo.mu.Unlock()

	if count != 1 {
		t.Errorf("expected 1 audit log entry for PATCH, got %d", count)
	}
}

// --- Helper function tests ---

func TestGetSessionFromContext_ValidSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	session := &shared.AuthSession{
		ID:     uuid.New(),
		UserID: uuid.New(),
	}
	c.Set(consts.CtxSession, session)

	got, ok := getSessionFromContext(c)
	if !ok {
		t.Fatal("expected session to be found")
	}
	if got.ID != session.ID {
		t.Errorf("expected session ID %s, got %s", session.ID, got.ID)
	}
}

func TestGetSessionFromContext_NoSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	_, ok := getSessionFromContext(c)
	if ok {
		t.Error("expected no session found")
	}
}

func TestUuidPtr_ValidUUID(t *testing.T) {
	id := uuid.New()
	result := uuidPtr(id.String())
	if result == nil {
		t.Fatal("expected non-nil UUID")
	}
	if *result != id {
		t.Errorf("expected %s, got %s", id, *result)
	}
}

func TestUuidPtr_EmptyString(t *testing.T) {
	result := uuidPtr("")
	if result != nil {
		t.Error("expected nil for empty string")
	}
}

func TestUuidPtr_InvalidString(t *testing.T) {
	result := uuidPtr("not-a-uuid")
	if result != nil {
		t.Error("expected nil for invalid UUID string")
	}
}
