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
	"gct/internal/platform/infrastructure/logger"
	"gct/internal/context/ops/systemerror/application/command"
	"gct/internal/context/ops/systemerror/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// --- Mock Logger ---

type sysErrMockLog struct {
	mu           sync.Mutex
	errorwCalled bool
	lastMsg      string
}

func (m *sysErrMockLog) Debug(args ...any)                                     {}
func (m *sysErrMockLog) Debugf(template string, args ...any)                   {}
func (m *sysErrMockLog) Debugw(msg string, keysAndValues ...any)               {}
func (m *sysErrMockLog) Info(args ...any)                                      {}
func (m *sysErrMockLog) Infof(template string, args ...any)                    {}
func (m *sysErrMockLog) Infow(msg string, keysAndValues ...any)                {}
func (m *sysErrMockLog) Warn(args ...any)                                      {}
func (m *sysErrMockLog) Warnf(template string, args ...any)                    {}
func (m *sysErrMockLog) Warnw(msg string, keysAndValues ...any)                {}
func (m *sysErrMockLog) Error(args ...any)                                     {}
func (m *sysErrMockLog) Errorf(template string, args ...any)                   {}
func (m *sysErrMockLog) Errorw(msg string, keysAndValues ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorwCalled = true
	m.lastMsg = msg
}
func (m *sysErrMockLog) Fatal(args ...any)                                     {}
func (m *sysErrMockLog) Fatalf(template string, args ...any)                   {}
func (m *sysErrMockLog) Fatalw(msg string, keysAndValues ...any)               {}
func (m *sysErrMockLog) Debugc(ctx context.Context, msg string, kv ...any)     {}
func (m *sysErrMockLog) Infoc(ctx context.Context, msg string, kv ...any)      {}
func (m *sysErrMockLog) Warnc(ctx context.Context, msg string, kv ...any)      {}
func (m *sysErrMockLog) Errorc(ctx context.Context, msg string, kv ...any)     {}
func (m *sysErrMockLog) Fatalc(ctx context.Context, msg string, kv ...any)     {}

var _ logger.Log = (*sysErrMockLog)(nil)

// --- Mock Repository ---

type mockSystemErrorRepo struct {
	mu       sync.Mutex
	saved    []*domain.SystemError
	saveErr  error
}

func (m *mockSystemErrorRepo) Save(ctx context.Context, entity *domain.SystemError) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saved = append(m.saved, entity)
	return m.saveErr
}

func (m *mockSystemErrorRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.SystemError, error) {
	return nil, nil
}

func (m *mockSystemErrorRepo) Update(ctx context.Context, entity *domain.SystemError) error {
	return nil
}

func (m *mockSystemErrorRepo) List(ctx context.Context, filter domain.SystemErrorFilter) ([]*domain.SystemError, int64, error) {
	return nil, 0, nil
}

// --- Mock EventBus ---

type mockEventBus struct{}

func (m *mockEventBus) Publish(_ context.Context, _ ...shared.DomainEvent) error { return nil }
func (m *mockEventBus) Subscribe(_ string, _ application.EventHandler) error     { return nil }

// --- Helper ---

func newSystemErrorMiddleware(repo *mockSystemErrorRepo, l logger.Log) *SystemErrorMiddleware {
	handler := command.NewCreateSystemErrorHandler(repo, &mockEventBus{}, l)
	return NewSystemErrorMiddleware(handler, l)
}

// --- Tests ---

func TestSystemErrorMiddleware_Recovery_PanicReturns500(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockSystemErrorRepo{}
	l := &sysErrMockLog{}
	mw := newSystemErrorMiddleware(repo, l)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(mw.Recovery())
	r.GET("/test", func(c *gin.Context) {
		panic("test panic")
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// Wait for async goroutine to persist
	time.Sleep(50 * time.Millisecond)

	l.mu.Lock()
	if !l.errorwCalled {
		t.Error("expected error to be logged on panic")
	}
	l.mu.Unlock()

	repo.mu.Lock()
	if len(repo.saved) == 0 {
		t.Error("expected system error to be persisted")
	}
	repo.mu.Unlock()
}

func TestSystemErrorMiddleware_Recovery_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockSystemErrorRepo{}
	l := &sysErrMockLog{}
	mw := newSystemErrorMiddleware(repo, l)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(mw.Recovery())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	time.Sleep(50 * time.Millisecond)

	repo.mu.Lock()
	if len(repo.saved) != 0 {
		t.Error("expected no system errors to be saved")
	}
	repo.mu.Unlock()
}

func TestSystemErrorMiddleware_Persist5xx_Saves5xxErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockSystemErrorRepo{}
	l := &sysErrMockLog{}
	mw := newSystemErrorMiddleware(repo, l)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(mw.Persist5xx())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something broke"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	time.Sleep(50 * time.Millisecond)

	repo.mu.Lock()
	if len(repo.saved) == 0 {
		t.Error("expected system error to be persisted for 500 status")
	}
	repo.mu.Unlock()
}

func TestSystemErrorMiddleware_Persist5xx_Ignores200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockSystemErrorRepo{}
	l := &sysErrMockLog{}
	mw := newSystemErrorMiddleware(repo, l)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(mw.Persist5xx())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	time.Sleep(50 * time.Millisecond)

	repo.mu.Lock()
	if len(repo.saved) != 0 {
		t.Error("expected no system errors for 200 response")
	}
	repo.mu.Unlock()
}

func TestSystemErrorMiddleware_Persist5xx_Ignores4xx(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockSystemErrorRepo{}
	l := &sysErrMockLog{}
	mw := newSystemErrorMiddleware(repo, l)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(mw.Persist5xx())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad input"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	time.Sleep(50 * time.Millisecond)

	repo.mu.Lock()
	if len(repo.saved) != 0 {
		t.Error("expected no system errors for 400 response")
	}
	repo.mu.Unlock()
}

func TestSystemErrorMiddleware_Persist5xx_CapturesGinErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockSystemErrorRepo{}
	l := &sysErrMockLog{}
	mw := newSystemErrorMiddleware(repo, l)

	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(mw.Persist5xx())
	r.GET("/test", func(c *gin.Context) {
		_ = c.Error(http.ErrBodyNotAllowed)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "fail"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	time.Sleep(50 * time.Millisecond)

	repo.mu.Lock()
	savedCount := len(repo.saved)
	repo.mu.Unlock()

	if savedCount == 0 {
		t.Error("expected at least one system error for gin error + 500")
	}
}
