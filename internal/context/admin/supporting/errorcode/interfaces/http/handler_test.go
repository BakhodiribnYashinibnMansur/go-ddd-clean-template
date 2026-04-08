package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/context/admin/supporting/errorcode"
	"gct/internal/context/admin/supporting/errorcode/application/command"
	"gct/internal/context/admin/supporting/errorcode/application/query"
	errcodeentity "gct/internal/context/admin/supporting/errorcode/domain/entity"
	errcoderepo "gct/internal/context/admin/supporting/errorcode/domain/repository"
	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"

	"gct/internal/kernel/outbox"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// --- Mocks ---

type mockRepo struct {
	saved   *errcodeentity.ErrorCode
	updated *errcodeentity.ErrorCode
	deleted errcodeentity.ErrorCodeID
	findFn  func(ctx context.Context, id errcodeentity.ErrorCodeID) (*errcodeentity.ErrorCode, error)
}

func (m *mockRepo) Save(_ context.Context, e *errcodeentity.ErrorCode) error {
	m.saved = e
	return nil
}
func (m *mockRepo) FindByID(ctx context.Context, id errcodeentity.ErrorCodeID) (*errcodeentity.ErrorCode, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, errcodeentity.ErrErrorCodeNotFound
}
func (m *mockRepo) Update(_ context.Context, e *errcodeentity.ErrorCode) error {
	m.updated = e
	return nil
}
func (m *mockRepo) Delete(_ context.Context, id errcodeentity.ErrorCodeID) error {
	m.deleted = id
	return nil
}

type mockReadRepo struct {
	view  *errcoderepo.ErrorCodeView
	views []*errcoderepo.ErrorCodeView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id errcodeentity.ErrorCodeID) (*errcoderepo.ErrorCodeView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, errcodeentity.ErrErrorCodeNotFound
}
func (m *mockReadRepo) List(_ context.Context, _ errcoderepo.ErrorCodeFilter) ([]*errcoderepo.ErrorCodeView, int64, error) {
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

	bc := &errorcode.BoundedContext{
		CreateErrorCode: command.NewCreateErrorCodeHandler(repo, outbox.NewEventCommitter(nil, nil, eb, l), l),
		UpdateErrorCode: command.NewUpdateErrorCodeHandler(repo, outbox.NewEventCommitter(nil, nil, eb, l), l),
		DeleteErrorCode: command.NewDeleteErrorCodeHandler(repo, outbox.NewEventCommitter(nil, nil, eb, l), l),
		GetErrorCode:    query.NewGetErrorCodeHandler(readRepo, l),
		ListErrorCodes:  query.NewListErrorCodesHandler(readRepo, l),
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

	body := CreateRequest{
		Code: "AUTH_001", Message: "unauthorized", HTTPStatus: 401,
		Category: "auth", Severity: "high",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/error-codes", bytes.NewBuffer(jsonBody))
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
	req, _ := http.NewRequest("POST", "/api/v1/error-codes", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_List_Success(t *testing.T) {
	t.Parallel()

	readRepo := &mockReadRepo{
		views: []*errcoderepo.ErrorCodeView{
			{ID: errcodeentity.NewErrorCodeID(), Code: "ERR_1", HTTPStatus: 400, Category: "c", Severity: "low", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		},
		total: 1,
	}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/error-codes?limit=10&offset=0", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_Success(t *testing.T) {
	t.Parallel()

	id := errcodeentity.NewErrorCodeID()
	readRepo := &mockReadRepo{
		view: &errcoderepo.ErrorCodeView{
			ID: id, Code: "ERR", HTTPStatus: 500, Category: "c", Severity: "s", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
	}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/error-codes/"+id.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_InvalidID(t *testing.T) {
	t.Parallel()

	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/error-codes/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Update_Success(t *testing.T) {
	t.Parallel()

	ec := errcodeentity.NewErrorCode("AUTH_001", "old", 401, "auth", "high", false, 0, "")
	repo := &mockRepo{
		findFn: func(_ context.Context, id errcodeentity.ErrorCodeID) (*errcodeentity.ErrorCode, error) {
			if id == ec.TypedID() {
				return ec, nil
			}
			return nil, errcodeentity.ErrErrorCodeNotFound
		},
	}
	router := setupRouter(repo, &mockReadRepo{})

	body := UpdateRequest{
		Message: "new msg", HTTPStatus: 403, Category: "auth", Severity: "critical",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/error-codes/"+ec.ID().String(), bytes.NewBuffer(jsonBody))
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
	req, _ := http.NewRequest("PATCH", "/api/v1/error-codes/bad-id", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Update_BadRequest(t *testing.T) {
	t.Parallel()

	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/error-codes/"+uuid.New().String(), bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Delete_Success(t *testing.T) {
	t.Parallel()

	ec := errcodeentity.NewErrorCode("ERR_DEL", "test", 500, "SYSTEM", "LOW", false, 0, "")
	repo := &mockRepo{
		findFn: func(_ context.Context, id errcodeentity.ErrorCodeID) (*errcodeentity.ErrorCode, error) {
			if id == ec.TypedID() {
				return ec, nil
			}
			return nil, errcodeentity.ErrErrorCodeNotFound
		},
	}
	router := setupRouter(repo, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/error-codes/"+ec.ID().String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if repo.deleted != ec.TypedID() {
		t.Errorf("expected deleted ID %s, got %s", ec.ID(), repo.deleted)
	}
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	t.Parallel()

	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/error-codes/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Get_NotFound(t *testing.T) {
	// mockReadRepo with nil view returns ErrErrorCodeNotFound for any UUID
	readRepo := &mockReadRepo{}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/error-codes/"+uuid.New().String(), nil)
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatalf("expected non-200 for not-found error code, got %d", w.Code)
	}
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	router := setupRouter(&mockRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/error-codes", bytes.NewBufferString(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestHandler_List_DefaultPagination(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*errcoderepo.ErrorCodeView{},
		total: 0,
	}
	router := setupRouter(&mockRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/error-codes", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for default pagination, got %d: %s", w.Code, w.Body.String())
	}
}
