package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/context/content/notification"
	"gct/internal/context/content/notification/application/command"
	"gct/internal/context/content/notification/application/query"
	"gct/internal/context/content/notification/domain"
	"gct/internal/platform/application"
	shared "gct/internal/platform/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// --- Mocks ---

type mockRepo struct {
	saved   *domain.Notification
	deleted uuid.UUID
}

func (m *mockRepo) Save(_ context.Context, n *domain.Notification) error {
	m.saved = n
	return nil
}
func (m *mockRepo) FindByID(_ context.Context, _ uuid.UUID) (*domain.Notification, error) {
	return nil, domain.ErrNotificationNotFound
}
func (m *mockRepo) Update(_ context.Context, _ *domain.Notification) error { return nil }
func (m *mockRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.deleted = id
	return nil
}

type mockReadRepo struct {
	view  *domain.NotificationView
	views []*domain.NotificationView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.NotificationView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrNotificationNotFound
}
func (m *mockReadRepo) List(_ context.Context, _ domain.NotificationFilter) ([]*domain.NotificationView, int64, error) {
	return m.views, m.total, nil
}

type mockEventBus struct{ published []shared.DomainEvent }

func (m *mockEventBus) Publish(_ context.Context, events ...shared.DomainEvent) error {
	m.published = append(m.published, events...)
	return nil
}
func (m *mockEventBus) Subscribe(_ string, _ application.EventHandler) error { return nil }

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

func setupRouter(repo *mockRepo, readRepo *mockReadRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)

	eb := &mockEventBus{}
	l := &mockLogger{}

	bc := &notification.BoundedContext{
		CreateNotification: command.NewCreateHandler(repo, eb, l),
		DeleteNotification: command.NewDeleteHandler(repo, eb, l),
		GetNotification:    query.NewGetHandler(readRepo, l),
		ListNotifications:  query.NewListHandler(readRepo, l),
	}

	r := gin.New()
	h := NewHandler(bc, l)
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)
	return r
}

// --- Tests ---

func TestHandler_Create_Success(t *testing.T) {
	repo := &mockRepo{}
	readRepo := &mockReadRepo{}
	router := setupRouter(repo, readRepo)

	userID := uuid.New()
	body := CreateRequest{UserID: userID, Title: "Alert", Message: "test msg", Type: "INFO"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/notifications", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Create_BadRequest(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/notifications", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_List_Success(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*domain.NotificationView{
			{ID: uuid.New(), UserID: uuid.New(), Title: "N1", Type: "INFO", CreatedAt: time.Now()},
		},
		total: 1,
	}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/notifications?limit=10&offset=0", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_Success(t *testing.T) {
	id := uuid.New()
	readRepo := &mockReadRepo{
		view: &domain.NotificationView{ID: id, UserID: uuid.New(), Title: "N", Type: "INFO", CreatedAt: time.Now()},
	}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/notifications/"+id.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_InvalidID(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/notifications/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Delete_Success(t *testing.T) {
	repo := &mockRepo{}
	router := setupRouter(repo, &mockReadRepo{})

	id := uuid.New()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/notifications/"+id.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/notifications/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
