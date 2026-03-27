package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/shared/application"
	shared "gct/internal/shared/domain"
	"gct/internal/webhook"
	"gct/internal/webhook/application/command"
	"gct/internal/webhook/application/query"
	"gct/internal/webhook/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Mock infrastructure
// ---------------------------------------------------------------------------

type mockWebhookRepo struct {
	savedWebhook   *domain.Webhook
	updatedWebhook *domain.Webhook
	deletedID      uuid.UUID
	findByIDFn     func(ctx context.Context, id uuid.UUID) (*domain.Webhook, error)
	saveFn         func(ctx context.Context, entity *domain.Webhook) error
	updateFn       func(ctx context.Context, entity *domain.Webhook) error
	deleteFn       func(ctx context.Context, id uuid.UUID) error
}

func (m *mockWebhookRepo) Save(ctx context.Context, entity *domain.Webhook) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, entity)
	}
	m.savedWebhook = entity
	return nil
}

func (m *mockWebhookRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Webhook, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, domain.ErrWebhookNotFound
}

func (m *mockWebhookRepo) Update(ctx context.Context, entity *domain.Webhook) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, entity)
	}
	m.updatedWebhook = entity
	return nil
}

func (m *mockWebhookRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	m.deletedID = id
	return nil
}

type mockEventBus struct {
	publishedEvents []shared.DomainEvent
}

func (m *mockEventBus) Publish(_ context.Context, events ...shared.DomainEvent) error {
	m.publishedEvents = append(m.publishedEvents, events...)
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

type mockReadRepo struct {
	view  *domain.WebhookView
	views []*domain.WebhookView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.WebhookView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrWebhookNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ domain.WebhookFilter) ([]*domain.WebhookView, int64, error) {
	return m.views, m.total, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func setupRouter(bc *webhook.BoundedContext) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(bc, &mockLogger{})
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)
	return r
}

func newBC(repo *mockWebhookRepo, readRepo *mockReadRepo) *webhook.BoundedContext {
	eb := &mockEventBus{}
	l := &mockLogger{}
	return &webhook.BoundedContext{
		CreateWebhook: command.NewCreateHandler(repo, eb, l),
		UpdateWebhook: command.NewUpdateHandler(repo, eb, l),
		DeleteWebhook: command.NewDeleteHandler(repo, eb, l),
		GetWebhook:    query.NewGetHandler(readRepo),
		ListWebhooks:  query.NewListHandler(readRepo),
	}
}

// ---------------------------------------------------------------------------
// Tests: POST /webhooks (Create)
// ---------------------------------------------------------------------------

func TestHandler_Create_Success(t *testing.T) {
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	body := CreateRequest{
		Name:    "my-hook",
		URL:     "https://example.com/hook",
		Secret:  "s3cret",
		Events:  []string{"user.created", "user.deleted"},
		Enabled: true,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/webhooks", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	if repo.savedWebhook == nil {
		t.Fatal("expected webhook to be saved")
	}
}

func TestHandler_Create_BadRequest(t *testing.T) {
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	// Missing required fields
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/webhooks", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/webhooks", bytes.NewBufferString(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestHandler_Create_MissingName(t *testing.T) {
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	body := `{"url":"https://example.com","secret":"sec","events":["a"]}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/webhooks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing name, got %d", w.Code)
	}
}

func TestHandler_Create_MissingURL(t *testing.T) {
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	body := `{"name":"hook","secret":"sec","events":["a"]}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/webhooks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing url, got %d", w.Code)
	}
}

func TestHandler_Create_MissingSecret(t *testing.T) {
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	body := `{"name":"hook","url":"https://example.com","events":["a"]}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/webhooks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing secret, got %d", w.Code)
	}
}

func TestHandler_Create_MissingEvents(t *testing.T) {
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	body := `{"name":"hook","url":"https://example.com","secret":"sec"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/webhooks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing events, got %d", w.Code)
	}
}

func TestHandler_Create_RepoError(t *testing.T) {
	repo := &mockWebhookRepo{
		saveFn: func(_ context.Context, _ *domain.Webhook) error {
			return errors.New("db connection failed")
		},
	}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	body := CreateRequest{
		Name:    "hook",
		URL:     "https://example.com",
		Secret:  "s3cret",
		Events:  []string{"user.created"},
		Enabled: true,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/webhooks", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandler_Create_ResponseFormat(t *testing.T) {
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	body := CreateRequest{
		Name:    "hook",
		URL:     "https://example.com",
		Secret:  "s3cret",
		Events:  []string{"user.created"},
		Enabled: true,
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/webhooks", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response should be valid JSON: %v", err)
	}
	if resp["success"] != true {
		t.Errorf("expected success true, got %v", resp["success"])
	}
}

// ---------------------------------------------------------------------------
// Tests: GET /webhooks (List)
// ---------------------------------------------------------------------------

func TestHandler_List_Success(t *testing.T) {
	now := time.Now()
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{
		views: []*domain.WebhookView{
			{ID: uuid.New(), Name: "hook-1", URL: "https://example.com/1", Events: []string{"a"}, Enabled: true, CreatedAt: now, UpdatedAt: now},
		},
		total: 1,
	}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/webhooks?limit=10&offset=0", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	total, ok := resp["total"].(float64)
	if !ok || total != 1 {
		t.Errorf("expected total 1, got %v", resp["total"])
	}
}

func TestHandler_List_Empty(t *testing.T) {
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{
		views: []*domain.WebhookView{},
		total: 0,
	}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/webhooks", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	total, ok := resp["total"].(float64)
	if !ok || total != 0 {
		t.Errorf("expected total 0, got %v", resp["total"])
	}
}

func TestHandler_List_DefaultPagination(t *testing.T) {
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{
		views: []*domain.WebhookView{},
		total: 0,
	}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	// No query params - should use defaults
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/webhooks", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_List_ResponseFormat(t *testing.T) {
	now := time.Now()
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{
		views: []*domain.WebhookView{
			{ID: uuid.New(), Name: "hook", URL: "https://example.com", Events: []string{}, Enabled: true, CreatedAt: now, UpdatedAt: now},
		},
		total: 1,
	}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/webhooks", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response should be valid JSON: %v", err)
	}
	if _, ok := resp["data"]; !ok {
		t.Error("response should contain 'data' field")
	}
	if _, ok := resp["total"]; !ok {
		t.Error("response should contain 'total' field")
	}
}

// ---------------------------------------------------------------------------
// Tests: GET /webhooks/:id (Get)
// ---------------------------------------------------------------------------

func TestHandler_Get_Success(t *testing.T) {
	webhookID := uuid.New()
	now := time.Now()
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{
		view: &domain.WebhookView{
			ID:        webhookID,
			Name:      "my-hook",
			URL:       "https://example.com/hook",
			Secret:    "s3cret",
			Events:    []string{"user.created"},
			Enabled:   true,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/webhooks/"+webhookID.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_InvalidID(t *testing.T) {
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/webhooks/not-a-uuid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Get_NotFound(t *testing.T) {
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{} // no view set
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/webhooks/"+uuid.New().String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandler_Get_ResponseFormat(t *testing.T) {
	webhookID := uuid.New()
	now := time.Now()
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{
		view: &domain.WebhookView{
			ID:        webhookID,
			Name:      "my-hook",
			URL:       "https://example.com",
			Events:    []string{},
			Enabled:   true,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/webhooks/"+webhookID.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response should be valid JSON: %v", err)
	}
	if _, ok := resp["data"]; !ok {
		t.Error("response should contain 'data' field")
	}
}

// ---------------------------------------------------------------------------
// Tests: PATCH /webhooks/:id (Update)
// ---------------------------------------------------------------------------

func TestHandler_Update_Success(t *testing.T) {
	existing := domain.NewWebhook("old-hook", "https://old.com", "oldsecret", []string{"a"}, true)
	existingID := existing.ID()

	repo := &mockWebhookRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Webhook, error) {
			if id == existingID {
				return existing, nil
			}
			return nil, domain.ErrWebhookNotFound
		},
	}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	newName := "new-hook"
	body := UpdateRequest{Name: &newName}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/webhooks/"+existingID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Update_InvalidID(t *testing.T) {
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/webhooks/not-a-uuid", bytes.NewBufferString(`{"name":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Update_InvalidJSON(t *testing.T) {
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/webhooks/"+uuid.New().String(), bytes.NewBufferString(`{bad json`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestHandler_Update_NotFound(t *testing.T) {
	repo := &mockWebhookRepo{} // default returns ErrWebhookNotFound
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	newName := "x"
	body := UpdateRequest{Name: &newName}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/webhooks/"+uuid.New().String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandler_Update_RepoError(t *testing.T) {
	existing := domain.NewWebhook("hook", "https://example.com", "secret", []string{"a"}, true)
	existingID := existing.ID()

	repo := &mockWebhookRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Webhook, error) {
			if id == existingID {
				return existing, nil
			}
			return nil, domain.ErrWebhookNotFound
		},
		updateFn: func(_ context.Context, _ *domain.Webhook) error {
			return errors.New("update failed")
		},
	}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	newName := "updated"
	body := UpdateRequest{Name: &newName}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/webhooks/"+existingID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandler_Update_PartialFields(t *testing.T) {
	existing := domain.NewWebhook("hook", "https://old.com", "oldsecret", []string{"a"}, true)
	existingID := existing.ID()

	repo := &mockWebhookRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Webhook, error) {
			if id == existingID {
				return existing, nil
			}
			return nil, domain.ErrWebhookNotFound
		},
	}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	disabled := false
	body := UpdateRequest{Enabled: &disabled}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/webhooks/"+existingID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Update_ResponseFormat(t *testing.T) {
	existing := domain.NewWebhook("hook", "https://example.com", "secret", []string{"a"}, true)
	existingID := existing.ID()

	repo := &mockWebhookRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Webhook, error) {
			if id == existingID {
				return existing, nil
			}
			return nil, domain.ErrWebhookNotFound
		},
	}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	newName := "updated"
	body := UpdateRequest{Name: &newName}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/webhooks/"+existingID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response should be valid JSON: %v", err)
	}
	if resp["success"] != true {
		t.Errorf("expected success true, got %v", resp["success"])
	}
}

// ---------------------------------------------------------------------------
// Tests: DELETE /webhooks/:id (Delete)
// ---------------------------------------------------------------------------

func TestHandler_Delete_Success(t *testing.T) {
	targetID := uuid.New()
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/webhooks/"+targetID.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/webhooks/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Delete_RepoError(t *testing.T) {
	repo := &mockWebhookRepo{
		deleteFn: func(_ context.Context, _ uuid.UUID) error {
			return errors.New("delete failed")
		},
	}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/webhooks/"+uuid.New().String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandler_Delete_ResponseFormat(t *testing.T) {
	repo := &mockWebhookRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/webhooks/"+uuid.New().String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response should be valid JSON: %v", err)
	}
	if resp["success"] != true {
		t.Errorf("expected success true, got %v", resp["success"])
	}
}
