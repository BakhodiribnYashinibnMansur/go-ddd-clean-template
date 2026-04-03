package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/session"
	appdto "gct/internal/session/application"
	"gct/internal/session/application/query"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// --- Mocks ---

type mockReadRepo struct {
	view    *appdto.SessionView
	views   []*appdto.SessionView
	total   int64
	findErr error
}

func (m *mockReadRepo) FindByID(_ context.Context, id uuid.UUID) (*appdto.SessionView, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, nil
}

func (m *mockReadRepo) List(_ context.Context, _ appdto.SessionsFilter) ([]*appdto.SessionView, int64, error) {
	return m.views, m.total, nil
}

type mockLogger struct{}

func (m *mockLogger) Debug(args ...any)                                          {}
func (m *mockLogger) Debugf(template string, args ...any)                        {}
func (m *mockLogger) Debugw(msg string, keysAndValues ...any)                    {}
func (m *mockLogger) Info(args ...any)                                           {}
func (m *mockLogger) Infof(template string, args ...any)                         {}
func (m *mockLogger) Infow(msg string, keysAndValues ...any)                     {}
func (m *mockLogger) Warn(args ...any)                                           {}
func (m *mockLogger) Warnf(template string, args ...any)                         {}
func (m *mockLogger) Warnw(msg string, keysAndValues ...any)                     {}
func (m *mockLogger) Error(args ...any)                                          {}
func (m *mockLogger) Errorf(template string, args ...any)                        {}
func (m *mockLogger) Errorw(msg string, keysAndValues ...any)                    {}
func (m *mockLogger) Fatal(args ...any)                                          {}
func (m *mockLogger) Fatalf(template string, args ...any)                        {}
func (m *mockLogger) Fatalw(msg string, keysAndValues ...any)                    {}
func (m *mockLogger) Debugc(_ context.Context, _ string, _ ...any)               {}
func (m *mockLogger) Infoc(_ context.Context, _ string, _ ...any)                {}
func (m *mockLogger) Warnc(_ context.Context, _ string, _ ...any)                {}
func (m *mockLogger) Errorc(_ context.Context, _ string, _ ...any)               {}
func (m *mockLogger) Fatalc(_ context.Context, _ string, _ ...any)               {}

// --- Helpers ---

func setupRouter(readRepo *mockReadRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)

	l := &mockLogger{}

	bc := &session.BoundedContext{
		GetSession:   query.NewGetSessionHandler(readRepo, l),
		ListSessions: query.NewListSessionsHandler(readRepo, l),
	}

	r := gin.New()
	h := NewHandler(bc, l)
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)
	return r
}

// --- Tests ---

func TestHandler_List_Success(t *testing.T) {
	now := time.Now()
	readRepo := &mockReadRepo{
		views: []*appdto.SessionView{
			{
				ID:           uuid.New(),
				UserID:       uuid.New(),
				DeviceID:     "device-1",
				DeviceName:   "Chrome on Mac",
				DeviceType:   "DESKTOP",
				IPAddress:    "192.168.1.1",
				UserAgent:    "Mozilla/5.0",
				ExpiresAt:    now.Add(7 * 24 * time.Hour),
				LastActivity: now,
				Revoked:      false,
				CreatedAt:    now,
			},
		},
		total: 1,
	}
	router := setupRouter(readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/sessions?limit=10&offset=0", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	readRepo := &mockReadRepo{
		view: &appdto.SessionView{
			ID:           id,
			UserID:       uuid.New(),
			DeviceID:     "device-1",
			DeviceName:   "Chrome on Mac",
			DeviceType:   "DESKTOP",
			IPAddress:    "192.168.1.1",
			UserAgent:    "Mozilla/5.0",
			ExpiresAt:    now.Add(7 * 24 * time.Hour),
			LastActivity: now,
			Revoked:      false,
			CreatedAt:    now,
		},
	}
	router := setupRouter(readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/sessions/"+id.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_InvalidID(t *testing.T) {
	router := setupRouter(&mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/sessions/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_List_WithUserIDFilter(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*appdto.SessionView{},
		total: 0,
	}
	router := setupRouter(readRepo)

	userID := uuid.New()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/sessions?user_id="+userID.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_List_InvalidUserID(t *testing.T) {
	router := setupRouter(&mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/sessions?user_id=bad-uuid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Get_NotFound(t *testing.T) {
	readRepo := &mockReadRepo{
		findErr: errors.New("session not found"),
	}
	router := setupRouter(readRepo)

	id := uuid.New()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/sessions/"+id.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatalf("expected non-200 when session not found, got %d", w.Code)
	}
}

func TestHandler_Get_InvalidUUID(t *testing.T) {
	router := setupRouter(&mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/sessions/not-a-uuid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_List_DefaultPagination(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*appdto.SessionView{},
		total: 0,
	}
	router := setupRouter(readRepo)

	w := httptest.NewRecorder()
	// No query params — should use default pagination and return 200
	req, _ := http.NewRequest("GET", "/api/v1/sessions", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
