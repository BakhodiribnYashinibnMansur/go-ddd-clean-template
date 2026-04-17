package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/context/ops/generic/ratelimit"
	"gct/internal/context/ops/generic/ratelimit/application/command"
	"gct/internal/context/ops/generic/ratelimit/application/query"
	ratelimitentity "gct/internal/context/ops/generic/ratelimit/domain/entity"
	ratelimitrepo "gct/internal/context/ops/generic/ratelimit/domain/repository"
	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"

	"gct/internal/kernel/outbox"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// --- Mocks ---

type mockRepo struct {
	saved   *ratelimitentity.RateLimit
	updated *ratelimitentity.RateLimit
	deleted ratelimitentity.RateLimitID
	findFn  func(ctx context.Context, id ratelimitentity.RateLimitID) (*ratelimitentity.RateLimit, error)
}

func (m *mockRepo) Save(_ context.Context, _ shared.Querier, e *ratelimitentity.RateLimit) error {
	m.saved = e
	return nil
}
func (m *mockRepo) FindByID(ctx context.Context, id ratelimitentity.RateLimitID) (*ratelimitentity.RateLimit, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, ratelimitentity.ErrRateLimitNotFound
}
func (m *mockRepo) Update(_ context.Context, _ shared.Querier, e *ratelimitentity.RateLimit) error {
	m.updated = e
	return nil
}
func (m *mockRepo) Delete(_ context.Context, _ shared.Querier, id ratelimitentity.RateLimitID) error {
	m.deleted = id
	return nil
}
func (m *mockRepo) List(_ context.Context, _ ratelimitrepo.RateLimitFilter) ([]*ratelimitentity.RateLimit, int64, error) {
	return nil, 0, nil
}

type mockReadRepo struct {
	view  *ratelimitrepo.RateLimitView
	views []*ratelimitrepo.RateLimitView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id ratelimitentity.RateLimitID) (*ratelimitrepo.RateLimitView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, ratelimitentity.ErrRateLimitNotFound
}
func (m *mockReadRepo) List(_ context.Context, _ ratelimitrepo.RateLimitFilter) ([]*ratelimitrepo.RateLimitView, int64, error) {
	return m.views, m.total, nil
}

type mockEventBus struct{ published []shared.DomainEvent }

func (m *mockEventBus) Publish(_ context.Context, events ...shared.DomainEvent) error {
	m.published = append(m.published, events...)
	return nil
}
func (m *mockEventBus) Subscribe(_ string, _ application.EventHandler) error { return nil }

type mockLogger struct{}

func (m *mockLogger) Debug(args ...any)                            {}
func (m *mockLogger) Debugf(template string, args ...any)          {}
func (m *mockLogger) Debugw(msg string, keysAndValues ...any)      {}
func (m *mockLogger) Info(args ...any)                             {}
func (m *mockLogger) Infof(template string, args ...any)           {}
func (m *mockLogger) Infow(msg string, keysAndValues ...any)       {}
func (m *mockLogger) Warn(args ...any)                             {}
func (m *mockLogger) Warnf(template string, args ...any)           {}
func (m *mockLogger) Warnw(msg string, keysAndValues ...any)       {}
func (m *mockLogger) Error(args ...any)                            {}
func (m *mockLogger) Errorf(template string, args ...any)          {}
func (m *mockLogger) Errorw(msg string, keysAndValues ...any)      {}
func (m *mockLogger) Fatal(args ...any)                            {}
func (m *mockLogger) Fatalf(template string, args ...any)          {}
func (m *mockLogger) Fatalw(msg string, keysAndValues ...any)      {}
func (m *mockLogger) Debugc(_ context.Context, _ string, _ ...any) {}
func (m *mockLogger) Infoc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Warnc(_ context.Context, _ string, _ ...any)  {}
func (m *mockLogger) Errorc(_ context.Context, _ string, _ ...any) {}
func (m *mockLogger) Fatalc(_ context.Context, _ string, _ ...any) {}

// --- Helpers ---

func setupRouter(repo *mockRepo, readRepo *mockReadRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)

	eb := &mockEventBus{}
	l := &mockLogger{}

	bc := &ratelimit.BoundedContext{
		CreateRateLimit: command.NewCreateRateLimitHandler(repo, outbox.NewEventCommitter(nil, nil, eb, l), l),
		UpdateRateLimit: command.NewUpdateRateLimitHandler(repo, outbox.NewEventCommitter(nil, nil, eb, l), l),
		DeleteRateLimit: command.NewDeleteRateLimitHandler(repo, outbox.NewEventCommitter(nil, nil, eb, l), l),
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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	readRepo := &mockReadRepo{
		views: []*ratelimitrepo.RateLimitView{
			{ID: ratelimitentity.NewRateLimitID(), Name: "r1", Rule: "/a", RequestsPerWindow: 10, WindowDuration: 30, Enabled: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
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
	t.Parallel()

	id := ratelimitentity.NewRateLimitID()
	readRepo := &mockReadRepo{
		view: &ratelimitrepo.RateLimitView{
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
	t.Parallel()

	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/rate-limits/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Update_Success(t *testing.T) {
	t.Parallel()

	rl := ratelimitentity.NewRateLimit("old", "/old", 10, 30, true)
	repo := &mockRepo{
		findFn: func(_ context.Context, id ratelimitentity.RateLimitID) (*ratelimitentity.RateLimit, error) {
			if id == rl.TypedID() {
				return rl, nil
			}
			return nil, ratelimitentity.ErrRateLimitNotFound
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
	t.Parallel()

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
	t.Parallel()

	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/rate-limits/"+uuid.New().String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	t.Parallel()

	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/rate-limits/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Get_NotFound(t *testing.T) {
	// mockReadRepo with nil view returns ErrRateLimitNotFound for any UUID
	readRepo := &mockReadRepo{}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/rate-limits/"+uuid.New().String(), nil)
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatalf("expected non-200 for not-found rate limit, got %d", w.Code)
	}
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/rate-limits", bytes.NewBufferString(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestHandler_List_DefaultPagination(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*ratelimitrepo.RateLimitView{},
		total: 0,
	}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/rate-limits", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for default pagination, got %d: %s", w.Code, w.Body.String())
	}
}
