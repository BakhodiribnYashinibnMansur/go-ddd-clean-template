package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/context/admin/featureflag"
	"gct/internal/context/admin/featureflag/application/command"
	"gct/internal/context/admin/featureflag/application/query"
	"gct/internal/context/admin/featureflag/domain"
	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// --- Mocks ---

type mockFeatureFlagRepo struct {
	saved   *domain.FeatureFlag
	updated *domain.FeatureFlag
	deleted bool
	findFn  func(ctx context.Context, id uuid.UUID) (*domain.FeatureFlag, error)
	listFn  func(ctx context.Context, f domain.FeatureFlagFilter) ([]*domain.FeatureFlag, int64, error)
}

func (m *mockFeatureFlagRepo) Save(_ context.Context, e *domain.FeatureFlag) error {
	m.saved = e
	return nil
}
func (m *mockFeatureFlagRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.FeatureFlag, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrFeatureFlagNotFound
}
func (m *mockFeatureFlagRepo) FindByKey(_ context.Context, _ string) (*domain.FeatureFlag, error) {
	return nil, domain.ErrFeatureFlagNotFound
}
func (m *mockFeatureFlagRepo) Update(_ context.Context, e *domain.FeatureFlag) error {
	m.updated = e
	return nil
}
func (m *mockFeatureFlagRepo) Delete(_ context.Context, _ uuid.UUID) error {
	m.deleted = true
	return nil
}
func (m *mockFeatureFlagRepo) FindAll(_ context.Context) ([]*domain.FeatureFlag, error) {
	return nil, nil
}

type mockRuleGroupRepo struct {
	saved   *domain.RuleGroup
	updated *domain.RuleGroup
	deleted bool
	findFn  func(ctx context.Context, id uuid.UUID) (*domain.RuleGroup, error)
}

func (m *mockRuleGroupRepo) Save(_ context.Context, e *domain.RuleGroup) error {
	m.saved = e
	return nil
}
func (m *mockRuleGroupRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.RuleGroup, error) {
	if m.findFn != nil {
		return m.findFn(ctx, id)
	}
	return nil, domain.ErrRuleGroupNotFound
}
func (m *mockRuleGroupRepo) Update(_ context.Context, e *domain.RuleGroup) error {
	m.updated = e
	return nil
}
func (m *mockRuleGroupRepo) Delete(_ context.Context, _ uuid.UUID) error {
	m.deleted = true
	return nil
}
func (m *mockRuleGroupRepo) FindByFlagID(_ context.Context, _ uuid.UUID) ([]*domain.RuleGroup, error) {
	return nil, nil
}
func (m *mockRuleGroupRepo) SaveCondition(_ context.Context, _ uuid.UUID, _ domain.Condition) error {
	return nil
}
func (m *mockRuleGroupRepo) DeleteConditionsByRuleGroupID(_ context.Context, _ uuid.UUID) error {
	return nil
}

type mockReadRepo struct {
	view  *domain.FeatureFlagView
	views []*domain.FeatureFlagView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.FeatureFlagView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrFeatureFlagNotFound
}
func (m *mockReadRepo) List(_ context.Context, _ domain.FeatureFlagFilter) ([]*domain.FeatureFlagView, int64, error) {
	return m.views, m.total, nil
}

type mockEventBus struct {
	published []shared.DomainEvent
}

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

type mockCachedEvaluator struct {
	results map[string]*query.EvalResult
}

func (m *mockCachedEvaluator) EvaluateFull(_ context.Context, key string, _ map[string]string) *query.EvalResult {
	if m.results == nil {
		return nil
	}
	return m.results[key]
}

// --- Helpers ---

func setupRouter(repo *mockFeatureFlagRepo, rgRepo *mockRuleGroupRepo, readRepo *mockReadRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)

	eb := &mockEventBus{}
	l := &mockLogger{}

	bc := &featureflag.BoundedContext{
		CreateFlag:      command.NewCreateHandler(repo, eb, l),
		UpdateFlag:      command.NewUpdateHandler(repo, eb, l),
		DeleteFlag:      command.NewDeleteHandler(repo, eb, l),
		CreateRuleGroup: command.NewCreateRuleGroupHandler(repo, rgRepo, eb, l),
		UpdateRuleGroup: command.NewUpdateRuleGroupHandler(rgRepo, eb, l),
		DeleteRuleGroup: command.NewDeleteRuleGroupHandler(rgRepo, eb, l),
		GetFlag:           query.NewGetHandler(readRepo, l),
		ListFlags:         query.NewListHandler(readRepo, l),
		EvaluateFlag:      query.NewEvaluateHandler(&mockCachedEvaluator{}),
		BatchEvaluateFlag: query.NewBatchEvaluateHandler(&mockCachedEvaluator{}),
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

	repo := &mockFeatureFlagRepo{}
	router := setupRouter(repo, &mockRuleGroupRepo{}, &mockReadRepo{})

	body := CreateRequest{Name: "Test Flag", Key: "test_flag", FlagType: "boolean"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/feature-flags", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if repo.saved == nil {
		t.Fatal("expected flag to be saved")
	}
}

func TestHandler_Create_BadRequest(t *testing.T) {
	t.Parallel()

	router := setupRouter(&mockFeatureFlagRepo{}, &mockRuleGroupRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/feature-flags", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_List_Success(t *testing.T) {
	t.Parallel()

	readRepo := &mockReadRepo{
		views: []*domain.FeatureFlagView{
			{ID: uuid.New(), Name: "Flag 1", Key: "flag_1"},
		},
		total: 1,
	}
	router := setupRouter(&mockFeatureFlagRepo{}, &mockRuleGroupRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/feature-flags?limit=10&offset=0", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_Success(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	readRepo := &mockReadRepo{
		view: &domain.FeatureFlagView{ID: id, Name: "Flag", Key: "flag"},
	}
	router := setupRouter(&mockFeatureFlagRepo{}, &mockRuleGroupRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/feature-flags/"+id.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_InvalidID(t *testing.T) {
	t.Parallel()

	router := setupRouter(&mockFeatureFlagRepo{}, &mockRuleGroupRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/feature-flags/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Delete_Success(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	repo := &mockFeatureFlagRepo{
		findFn: func(_ context.Context, fid uuid.UUID) (*domain.FeatureFlag, error) {
			now := time.Now()
			return domain.ReconstructFeatureFlag(fid, now, now, nil, "flag", "flag_key", "", "boolean", "false", 0, true, nil), nil
		},
	}
	router := setupRouter(repo, &mockRuleGroupRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/feature-flags/"+id.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	t.Parallel()

	router := setupRouter(&mockFeatureFlagRepo{}, &mockRuleGroupRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/feature-flags/not-a-uuid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Update_BadRequest(t *testing.T) {
	t.Parallel()

	router := setupRouter(&mockFeatureFlagRepo{}, &mockRuleGroupRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/feature-flags/bad-id", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_CreateRuleGroup_Success(t *testing.T) {
	t.Parallel()

	flagID := uuid.New()
	repo := &mockFeatureFlagRepo{
		findFn: func(_ context.Context, id uuid.UUID) (*domain.FeatureFlag, error) {
			now := time.Now()
			return domain.ReconstructFeatureFlag(id, now, now, nil, "flag", "key", "", "boolean", "false", 0, true, nil), nil
		},
	}
	rgRepo := &mockRuleGroupRepo{}
	router := setupRouter(repo, rgRepo, &mockReadRepo{})

	body := CreateRuleGroupRequest{
		Name:      "Beta Users",
		Variation: "true",
		Priority:  1,
		Conditions: []ConditionRequest{
			{Attribute: "email", Operator: "contains", Value: "@beta.com"},
		},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/feature-flags/"+flagID.String()+"/rule-groups", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_CreateRuleGroup_InvalidFlagID(t *testing.T) {
	t.Parallel()

	router := setupRouter(&mockFeatureFlagRepo{}, &mockRuleGroupRepo{}, &mockReadRepo{})

	body := CreateRuleGroupRequest{Name: "Test", Variation: "true"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/feature-flags/bad-id/rule-groups", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_DeleteRuleGroup_InvalidID(t *testing.T) {
	t.Parallel()

	router := setupRouter(&mockFeatureFlagRepo{}, &mockRuleGroupRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/feature-flags/"+uuid.New().String()+"/rule-groups/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_UpdateRuleGroup_InvalidID(t *testing.T) {
	t.Parallel()

	router := setupRouter(&mockFeatureFlagRepo{}, &mockRuleGroupRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/feature-flags/"+uuid.New().String()+"/rule-groups/bad-id", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func setupRouterWithEvaluator(evalResults map[string]*query.EvalResult) *gin.Engine {
	gin.SetMode(gin.TestMode)

	l := &mockLogger{}
	eval := &mockCachedEvaluator{results: evalResults}

	bc := &featureflag.BoundedContext{
		CreateFlag:        command.NewCreateHandler(&mockFeatureFlagRepo{}, &mockEventBus{}, l),
		UpdateFlag:        command.NewUpdateHandler(&mockFeatureFlagRepo{}, &mockEventBus{}, l),
		DeleteFlag:        command.NewDeleteHandler(&mockFeatureFlagRepo{}, &mockEventBus{}, l),
		CreateRuleGroup:   command.NewCreateRuleGroupHandler(&mockFeatureFlagRepo{}, &mockRuleGroupRepo{}, &mockEventBus{}, l),
		UpdateRuleGroup:   command.NewUpdateRuleGroupHandler(&mockRuleGroupRepo{}, &mockEventBus{}, l),
		DeleteRuleGroup:   command.NewDeleteRuleGroupHandler(&mockRuleGroupRepo{}, &mockEventBus{}, l),
		GetFlag:           query.NewGetHandler(&mockReadRepo{}, l),
		ListFlags:         query.NewListHandler(&mockReadRepo{}, l),
		EvaluateFlag:      query.NewEvaluateHandler(eval),
		BatchEvaluateFlag: query.NewBatchEvaluateHandler(eval),
	}

	r := gin.New()
	h := NewHandler(bc, l)
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)
	return r
}

func TestHandler_Evaluate_Success(t *testing.T) {
	t.Parallel()

	router := setupRouterWithEvaluator(map[string]*query.EvalResult{
		"dark_mode": {Value: "true", FlagType: "bool"},
	})

	body := EvaluateRequest{Key: "dark_mode", UserAttrs: map[string]string{"platform": "web"}}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/feature-flags/evaluate", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["key"] != "dark_mode" {
		t.Errorf("expected key dark_mode, got %v", resp["key"])
	}
	if resp["value"] != "true" {
		t.Errorf("expected value true, got %v", resp["value"])
	}
}

func TestHandler_Evaluate_NotFound(t *testing.T) {
	t.Parallel()

	router := setupRouterWithEvaluator(nil)

	body := EvaluateRequest{Key: "nonexistent", UserAttrs: nil}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/feature-flags/evaluate", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Evaluate_BadRequest(t *testing.T) {
	t.Parallel()

	router := setupRouterWithEvaluator(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/feature-flags/evaluate", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_BatchEvaluate_Success(t *testing.T) {
	t.Parallel()

	router := setupRouterWithEvaluator(map[string]*query.EvalResult{
		"flag_a": {Value: "true", FlagType: "bool"},
		"flag_b": {Value: "dark", FlagType: "string"},
	})

	body := BatchEvaluateRequest{
		Keys:      []string{"flag_a", "flag_b", "flag_missing"},
		UserAttrs: map[string]string{"platform": "web"},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/feature-flags/evaluate/batch", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	flags := resp["flags"].(map[string]any)
	if len(flags) != 2 {
		t.Errorf("expected 2 flags, got %d", len(flags))
	}
}

func TestHandler_BatchEvaluate_BadRequest(t *testing.T) {
	t.Parallel()

	router := setupRouterWithEvaluator(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/feature-flags/evaluate/batch", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Additional error-path, parsing, and pagination tests ---

func TestHandler_GetFlag_NotFound(t *testing.T) {
	// mockReadRepo with no view set returns ErrFeatureFlagNotFound
	readRepo := &mockReadRepo{}
	router := setupRouter(&mockFeatureFlagRepo{}, &mockRuleGroupRepo{}, readRepo)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/feature-flags/"+uuid.New().String(), nil)
	router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatalf("expected non-200 for not-found flag, got %d", w.Code)
	}
}

func TestHandler_ListFlags_DefaultPagination(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*domain.FeatureFlagView{},
		total: 0,
	}
	router := setupRouter(&mockFeatureFlagRepo{}, &mockRuleGroupRepo{}, readRepo)

	// No query params — should use default pagination and succeed
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/feature-flags", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for default pagination, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_CreateFlag_EmptyBody(t *testing.T) {
	router := setupRouter(&mockFeatureFlagRepo{}, &mockRuleGroupRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/feature-flags", bytes.NewBufferString(``))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty body, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_CreateFlag_InvalidJSON(t *testing.T) {
	router := setupRouter(&mockFeatureFlagRepo{}, &mockRuleGroupRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/feature-flags", bytes.NewBufferString(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_DeleteFlag_InvalidUUID(t *testing.T) {
	router := setupRouter(&mockFeatureFlagRepo{}, &mockRuleGroupRepo{}, &mockReadRepo{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/feature-flags/not-a-uuid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid UUID, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Evaluate_EmptyBody(t *testing.T) {
	router := setupRouterWithEvaluator(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/feature-flags/evaluate", bytes.NewBufferString(``))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty evaluate body, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_BatchEvaluate_EmptyBody(t *testing.T) {
	router := setupRouterWithEvaluator(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/feature-flags/evaluate/batch", bytes.NewBufferString(``))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty batch body, got %d: %s", w.Code, w.Body.String())
	}
}
