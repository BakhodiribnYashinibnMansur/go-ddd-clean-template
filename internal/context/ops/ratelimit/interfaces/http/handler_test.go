package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/context/ops/ratelimit"
	"gct/internal/context/ops/ratelimit/application/command"
	"gct/internal/context/ops/ratelimit/application/query"
	"gct/internal/context/ops/ratelimit/domain"
	"gct/internal/platform/application"
	shared "gct/internal/platform/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// --- Mocks ---

type mockRepo struct {
	saved   *domain.RateLimit
	updated *domain.RateLimit
	deleted uuid.UUID
	findFn  func(ctx context.Context, id uuid.UUID) (*domain.RateLimit, error)
}

func (m *mockRepo) Save(_ context.Context, e *domain.RateLimit) error {
	m.saved = e
	return nil
}
func (m *mockRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.RateLimit, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrRateLimitNotFound
}
func (m *mockRepo) Update(_ context.Context, e *domain.RateLimit) error {
	m.updated = e
	return nil
}
func (m *mockRepo) Delete(_ context.Context, id uuid.UUID) error {
	m.deleted = id
	return nil
}
func (m *mockRepo) List(_ context.Context, _ domain.RateLimitFilter) ([]*domain.RateLimit, int64, error) {
	return nil, 0, nil
}

type mockReadRepo struct {
	view  *domain.RateLimitView
	views []*domain.RateLimitView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.RateLimitView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrRateLimitNotFound
}
func (m *mockReadRepo) List(_ context.Context, _ domain.RateLimitFilter) ([]*domain.RateLimitView, int64, error) {
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

	bc := &ratelimit.BoundedContext{
		CreateRateLimit: command.NewCreateRateLimitHandler(repo, eb, l),
		UpdateRateLimit: command.NewUpdateRateLimitHandler(repo, eb, l),
		DeleteRateLimit: command.NewDeleteRateLimitHandler(repo, l),
		GetRateLimit:    query.NewGetRateLimitHandler(readRepo, l),
		ListRateLimits:  query.NewListRateLimitsHandler(readRepo, l),
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
	router := setupRouter(repo, &mockReadRepo{})

	body := CreateRequest{
		Name:              "api-global",
		Rule:              "/api/*",
		RequestsPerWindow: 100,
		WindowDuration:    60,
		Enabled:           true,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/rate-limits", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Create_BadRequest(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/rate-limits", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_List_Success(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*domain.RateLimitView{
			{ID: uuid.New(), Name: "r1", Rule: "/a", RequestsPerWindow: 10, WindowDuration: 30, Enabled: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
		total: 1,
	}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/rate-limits?limit=10&offset=0", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_Success(t *testing.T) {
	id := uuid.New()
	readRepo := &mockReadRepo{
		view: &domain.RateLimitView{
			ID: id, Name: "r1", Rule: "/a", RequestsPerWindow: 10, WindowDuration: 30, Enabled: true, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
	}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/rate-limits/"+id.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_InvalidID(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/rate-limits/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Update_Success(t *testing.T) {
	rl := domain.NewRateLimit("old", "/old", 10, 30, true)
	repo := &mockRepo{
		findFn: func(_ context.Context, id uuid.UUID) (*domain.RateLimit, error) {
			if id == rl.ID() {
				return rl, nil
			}
			return nil, domain.ErrRateLimitNotFound
		},
	}
	router := setupRouter(repo, &mockReadRepo{})

	newName := "updated"
	body := UpdateRequest{Name: &newName}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/rate-limits/"+rl.ID().String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Update_InvalidID(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/rate-limits/bad-id", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Delete_Success(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/rate-limits/"+uuid.New().String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/rate-limits/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
