package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/context/content/generic/notification"
	"gct/internal/context/content/generic/notification/application/command"
	"gct/internal/context/content/generic/notification/application/query"
	notifentity "gct/internal/context/content/generic/notification/domain/entity"
	notifrepo "gct/internal/context/content/generic/notification/domain/repository"
	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// --- Mocks ---

type mockRepo struct {
	saved   *notifentity.Notification
	deleted notifentity.NotificationID
}

func (m *mockRepo) Save(_ context.Context, n *notifentity.Notification) error {
	m.saved = n
	return nil
}
func (m *mockRepo) FindByID(_ context.Context, _ notifentity.NotificationID) (*notifentity.Notification, error) {
	return nil, notifentity.ErrNotificationNotFound
}
func (m *mockRepo) Update(_ context.Context, _ *notifentity.Notification) error { return nil }
func (m *mockRepo) Delete(_ context.Context, id notifentity.NotificationID) error {
	m.deleted = id
	return nil
}

type mockReadRepo struct {
	view  *notifrepo.NotificationView
	views []*notifrepo.NotificationView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id notifentity.NotificationID) (*notifrepo.NotificationView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, notifentity.ErrNotificationNotFound
}
func (m *mockReadRepo) List(_ context.Context, _ notifrepo.NotificationFilter) ([]*notifrepo.NotificationView, int64, error) {
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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	readRepo := &mockReadRepo{
		views: []*notifrepo.NotificationView{
			{ID: notifentity.NewNotificationID(), UserID: uuid.New(), Title: "N1", Type: "INFO", CreatedAt: time.Now()},
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
	t.Parallel()

	id := notifentity.NewNotificationID()
	readRepo := &mockReadRepo{
		view: &notifrepo.NotificationView{ID: id, UserID: uuid.New(), Title: "N", Type: "INFO", CreatedAt: time.Now()},
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
	t.Parallel()

	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/notifications/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Delete_Success(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/notifications/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Get_NotFound(t *testing.T) {
	// readRepo has no view set, so FindByID returns ErrNotificationNotFound
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	id := uuid.New()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/notifications/"+id.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatalf("expected non-200 for missing notification, got %d", w.Code)
	}
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/notifications", bytes.NewBufferString(`not json at all`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestHandler_List_DefaultPagination(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*notifrepo.NotificationView{},
		total: 0,
	}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/notifications", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with default pagination, got %d: %s", w.Code, w.Body.String())
	}
}
