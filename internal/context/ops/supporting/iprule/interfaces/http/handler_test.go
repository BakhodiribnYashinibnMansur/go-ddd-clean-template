package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/context/ops/supporting/iprule"
	"gct/internal/context/ops/supporting/iprule/application/command"
	"gct/internal/context/ops/supporting/iprule/application/query"
	ipruleentity "gct/internal/context/ops/supporting/iprule/domain/entity"
	iprulerepo "gct/internal/context/ops/supporting/iprule/domain/repository"
	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"

	"gct/internal/kernel/outbox"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// --- Mocks ---

type mockRepo struct {
	saved   *ipruleentity.IPRule
	updated *ipruleentity.IPRule
	deleted ipruleentity.IPRuleID
	findFn  func(ctx context.Context, id ipruleentity.IPRuleID) (*ipruleentity.IPRule, error)
}

func (m *mockRepo) Save(_ context.Context, e *ipruleentity.IPRule) error {
	m.saved = e
	return nil
}
func (m *mockRepo) FindByID(ctx context.Context, id ipruleentity.IPRuleID) (*ipruleentity.IPRule, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, ipruleentity.ErrIPRuleNotFound
}
func (m *mockRepo) Update(_ context.Context, e *ipruleentity.IPRule) error {
	m.updated = e
	return nil
}
func (m *mockRepo) Delete(_ context.Context, id ipruleentity.IPRuleID) error {
	m.deleted = id
	return nil
}
func (m *mockRepo) List(_ context.Context, _ iprulerepo.IPRuleFilter) ([]*ipruleentity.IPRule, int64, error) {
	return nil, 0, nil
}

type mockReadRepo struct {
	view  *iprulerepo.IPRuleView
	views []*iprulerepo.IPRuleView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id ipruleentity.IPRuleID) (*iprulerepo.IPRuleView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, ipruleentity.ErrIPRuleNotFound
}
func (m *mockReadRepo) List(_ context.Context, _ iprulerepo.IPRuleFilter) ([]*iprulerepo.IPRuleView, int64, error) {
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

	bc := &iprule.BoundedContext{
		CreateIPRule: command.NewCreateIPRuleHandler(repo, outbox.NewEventCommitter(nil, nil, eb, l), l),
		UpdateIPRule: command.NewUpdateIPRuleHandler(repo, outbox.NewEventCommitter(nil, nil, eb, l), l),
		DeleteIPRule: command.NewDeleteIPRuleHandler(repo, l),
		GetIPRule:    query.NewGetIPRuleHandler(readRepo, l),
		ListIPRules:  query.NewListIPRulesHandler(readRepo, l),
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
		IPAddress: "192.168.1.100",
		Action:    "DENY",
		Reason:    "suspicious",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ip-rules", bytes.NewBuffer(jsonBody))
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
	req, _ := http.NewRequest("POST", "/api/v1/ip-rules", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_List_Success(t *testing.T) {
	t.Parallel()

	readRepo := &mockReadRepo{
		views: []*iprulerepo.IPRuleView{
			{ID: ipruleentity.NewIPRuleID(), IPAddress: "1.1.1.1", Action: "DENY", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
		total: 1,
	}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ip-rules?limit=10&offset=0", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_Success(t *testing.T) {
	t.Parallel()

	id := ipruleentity.NewIPRuleID()
	readRepo := &mockReadRepo{
		view: &iprulerepo.IPRuleView{
			ID: id, IPAddress: "1.1.1.1", Action: "DENY", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
	}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ip-rules/"+id.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_InvalidID(t *testing.T) {
	t.Parallel()

	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ip-rules/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Update_Success(t *testing.T) {
	t.Parallel()

	r := ipruleentity.NewIPRule("192.168.1.1", "DENY", "test", nil)
	repo := &mockRepo{
		findFn: func(_ context.Context, id ipruleentity.IPRuleID) (*ipruleentity.IPRule, error) {
			if id == r.TypedID() {
				return r, nil
			}
			return nil, ipruleentity.ErrIPRuleNotFound
		},
	}
	router := setupRouter(repo, &mockReadRepo{})

	newIP := "10.0.0.1"
	body := UpdateRequest{IPAddress: &newIP}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/ip-rules/"+r.ID().String(), bytes.NewBuffer(jsonBody))
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
	req, _ := http.NewRequest("PATCH", "/api/v1/ip-rules/bad-id", bytes.NewBufferString(`{}`))
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
	req, _ := http.NewRequest("DELETE", "/api/v1/ip-rules/"+uuid.New().String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	t.Parallel()

	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/ip-rules/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Get_NotFound(t *testing.T) {
	// mockReadRepo with nil view returns ErrIPRuleNotFound for any UUID
	readRepo := &mockReadRepo{}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ip-rules/"+uuid.New().String(), nil)
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatalf("expected non-200 for not-found IP rule, got %d", w.Code)
	}
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ip-rules", bytes.NewBufferString(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestHandler_List_DefaultPagination(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*iprulerepo.IPRuleView{},
		total: 0,
	}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ip-rules", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for default pagination, got %d: %s", w.Code, w.Body.String())
	}
}
