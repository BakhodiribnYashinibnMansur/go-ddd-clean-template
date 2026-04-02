# Feature Flag Evaluate Endpoint Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add evaluate and batch-evaluate HTTP endpoints so clients can resolve feature flag values using `userAttrs` (including `platform`).

**Architecture:** Two new CQRS query handlers (`EvaluateHandler`, `BatchEvaluateHandler`) backed by existing `CachedEvaluator`. A new `EvaluateFull` method on `CachedEvaluator` returns both value and flag type. HTTP handler methods wire requests to query handlers.

**Tech Stack:** Go, Gin, pgxutil (OTel tracing), CachedEvaluator (sync.Map cache)

---

### Task 1: Add `EvaluateFull` method to CachedEvaluator

**Files:**
- Modify: `internal/featureflag/infrastructure/cache/evaluator_cache.go`

- [ ] **Step 1: Add `EvalResult` struct and `EvaluateFull` method**

```go
// EvalResult holds the evaluated value and the flag's type.
type EvalResult struct {
	Value    string
	FlagType string
}

// EvaluateFull evaluates a flag and returns the value together with its type.
// Returns nil when the flag does not exist.
func (ce *CachedEvaluator) EvaluateFull(ctx context.Context, key string, userAttrs map[string]string) *EvalResult {
	ff := ce.getFlag(ctx, key)
	if ff == nil {
		return nil
	}
	return &EvalResult{
		Value:    ff.Evaluate(userAttrs),
		FlagType: ff.FlagType(),
	}
}
```

Add this after the existing `GetFloat` method (after line 89), before `getFlag`.

- [ ] **Step 2: Run existing tests to verify nothing broke**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/featureflag/infrastructure/cache/... -v`
Expected: All existing tests PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/featureflag/infrastructure/cache/evaluator_cache.go
git commit -m "feat(featureflag): add EvaluateFull method to CachedEvaluator"
```

---

### Task 2: Create `EvaluateHandler` query handler

**Files:**
- Create: `internal/featureflag/application/query/evaluate.go`
- Create: `internal/featureflag/application/query/evaluate_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/featureflag/application/query/evaluate_test.go`:

```go
package query_test

import (
	"context"
	"testing"

	"gct/internal/featureflag/application/query"
	"gct/internal/featureflag/infrastructure/cache"
)

type mockEvaluator struct {
	result *cache.EvalResult
}

func (m *mockEvaluator) EvaluateFull(_ context.Context, _ string, _ map[string]string) *cache.EvalResult {
	return m.result
}

func TestEvaluateHandler_ReturnsValue(t *testing.T) {
	eval := &mockEvaluator{result: &cache.EvalResult{Value: "true", FlagType: "bool"}}
	h := query.NewEvaluateHandler(eval)

	result, err := h.Handle(context.Background(), query.EvaluateQuery{
		Key:       "dark_mode",
		UserAttrs: map[string]string{"platform": "web"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Key != "dark_mode" {
		t.Errorf("expected key dark_mode, got %s", result.Key)
	}
	if result.Value != "true" {
		t.Errorf("expected value true, got %s", result.Value)
	}
	if result.FlagType != "bool" {
		t.Errorf("expected flag_type bool, got %s", result.FlagType)
	}
}

func TestEvaluateHandler_FlagNotFound(t *testing.T) {
	eval := &mockEvaluator{result: nil}
	h := query.NewEvaluateHandler(eval)

	_, err := h.Handle(context.Background(), query.EvaluateQuery{
		Key:       "nonexistent",
		UserAttrs: nil,
	})
	if err == nil {
		t.Fatal("expected error for missing flag")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/featureflag/application/query/... -run TestEvaluateHandler -v`
Expected: FAIL — `query.NewEvaluateHandler` not defined.

- [ ] **Step 3: Write the implementation**

Create `internal/featureflag/application/query/evaluate.go`:

```go
package query

import (
	"context"
	"errors"

	"gct/internal/featureflag/infrastructure/cache"
	"gct/internal/shared/infrastructure/pgxutil"
)

// FlagEvaluator is the interface the evaluate handler needs from the cache layer.
type FlagEvaluator interface {
	EvaluateFull(ctx context.Context, key string, userAttrs map[string]string) *cache.EvalResult
}

// EvaluateQuery holds the input for evaluating a single feature flag.
type EvaluateQuery struct {
	Key       string
	UserAttrs map[string]string
}

// EvaluateResult holds the output of a single flag evaluation.
type EvaluateResult struct {
	Key      string
	Value    string
	FlagType string
}

// EvaluateHandler handles the EvaluateQuery.
type EvaluateHandler struct {
	evaluator FlagEvaluator
}

// NewEvaluateHandler creates a new EvaluateHandler.
func NewEvaluateHandler(evaluator FlagEvaluator) *EvaluateHandler {
	return &EvaluateHandler{evaluator: evaluator}
}

// Handle evaluates a single feature flag for the given user attributes.
func (h *EvaluateHandler) Handle(ctx context.Context, q EvaluateQuery) (_ *EvaluateResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "EvaluateHandler.Handle")
	defer func() { end(err) }()

	result := h.evaluator.EvaluateFull(ctx, q.Key, q.UserAttrs)
	if result == nil {
		return nil, errors.New("feature flag not found")
	}

	return &EvaluateResult{
		Key:      q.Key,
		Value:    result.Value,
		FlagType: result.FlagType,
	}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/featureflag/application/query/... -run TestEvaluateHandler -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/featureflag/application/query/evaluate.go internal/featureflag/application/query/evaluate_test.go
git commit -m "feat(featureflag): add EvaluateHandler query handler"
```

---

### Task 3: Create `BatchEvaluateHandler` query handler

**Files:**
- Create: `internal/featureflag/application/query/evaluate_batch.go`
- Create: `internal/featureflag/application/query/evaluate_batch_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/featureflag/application/query/evaluate_batch_test.go`:

```go
package query_test

import (
	"context"
	"testing"

	"gct/internal/featureflag/application/query"
	"gct/internal/featureflag/infrastructure/cache"
)

type batchMockEvaluator struct {
	results map[string]*cache.EvalResult
}

func (m *batchMockEvaluator) EvaluateFull(_ context.Context, key string, _ map[string]string) *cache.EvalResult {
	return m.results[key]
}

func TestBatchEvaluateHandler_ReturnsMultiple(t *testing.T) {
	eval := &batchMockEvaluator{results: map[string]*cache.EvalResult{
		"flag_a": {Value: "true", FlagType: "bool"},
		"flag_b": {Value: "dark", FlagType: "string"},
	}}
	h := query.NewBatchEvaluateHandler(eval)

	result, err := h.Handle(context.Background(), query.BatchEvaluateQuery{
		Keys:      []string{"flag_a", "flag_b", "flag_missing"},
		UserAttrs: map[string]string{"platform": "web"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Flags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(result.Flags))
	}
	if result.Flags["flag_a"].Value != "true" {
		t.Errorf("expected flag_a value true, got %s", result.Flags["flag_a"].Value)
	}
	if result.Flags["flag_b"].FlagType != "string" {
		t.Errorf("expected flag_b type string, got %s", result.Flags["flag_b"].FlagType)
	}
}

func TestBatchEvaluateHandler_AllMissing(t *testing.T) {
	eval := &batchMockEvaluator{results: map[string]*cache.EvalResult{}}
	h := query.NewBatchEvaluateHandler(eval)

	result, err := h.Handle(context.Background(), query.BatchEvaluateQuery{
		Keys:      []string{"missing_a", "missing_b"},
		UserAttrs: nil,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Flags) != 0 {
		t.Fatalf("expected 0 flags, got %d", len(result.Flags))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/featureflag/application/query/... -run TestBatchEvaluateHandler -v`
Expected: FAIL — `query.NewBatchEvaluateHandler` not defined.

- [ ] **Step 3: Write the implementation**

Create `internal/featureflag/application/query/evaluate_batch.go`:

```go
package query

import (
	"context"

	"gct/internal/shared/infrastructure/pgxutil"
)

// BatchEvaluateQuery holds the input for evaluating multiple feature flags.
type BatchEvaluateQuery struct {
	Keys      []string
	UserAttrs map[string]string
}

// BatchEvaluateResult holds the output of a batch flag evaluation.
type BatchEvaluateResult struct {
	Flags map[string]EvaluateResult
}

// BatchEvaluateHandler handles the BatchEvaluateQuery.
type BatchEvaluateHandler struct {
	evaluator FlagEvaluator
}

// NewBatchEvaluateHandler creates a new BatchEvaluateHandler.
func NewBatchEvaluateHandler(evaluator FlagEvaluator) *BatchEvaluateHandler {
	return &BatchEvaluateHandler{evaluator: evaluator}
}

// Handle evaluates multiple feature flags for the given user attributes.
// Flags that do not exist are omitted from the result.
func (h *BatchEvaluateHandler) Handle(ctx context.Context, q BatchEvaluateQuery) (_ *BatchEvaluateResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "BatchEvaluateHandler.Handle")
	defer func() { end(err) }()

	flags := make(map[string]EvaluateResult, len(q.Keys))
	for _, key := range q.Keys {
		result := h.evaluator.EvaluateFull(ctx, key, q.UserAttrs)
		if result == nil {
			continue
		}
		flags[key] = EvaluateResult{
			Key:      key,
			Value:    result.Value,
			FlagType: result.FlagType,
		}
	}

	return &BatchEvaluateResult{Flags: flags}, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/featureflag/application/query/... -run TestBatchEvaluateHandler -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/featureflag/application/query/evaluate_batch.go internal/featureflag/application/query/evaluate_batch_test.go
git commit -m "feat(featureflag): add BatchEvaluateHandler query handler"
```

---

### Task 4: Add request structs, handler methods, and routes

**Files:**
- Modify: `internal/featureflag/interfaces/http/request.go`
- Modify: `internal/featureflag/interfaces/http/handler.go`
- Modify: `internal/featureflag/interfaces/http/routes.go`

- [ ] **Step 1: Add request structs to `request.go`**

Append to end of `internal/featureflag/interfaces/http/request.go`:

```go
// EvaluateRequest represents the request body for evaluating a single feature flag.
type EvaluateRequest struct {
	Key       string            `json:"key" binding:"required"`
	UserAttrs map[string]string `json:"user_attrs"`
}

// BatchEvaluateRequest represents the request body for evaluating multiple feature flags.
type BatchEvaluateRequest struct {
	Keys      []string          `json:"keys" binding:"required,min=1"`
	UserAttrs map[string]string `json:"user_attrs"`
}
```

- [ ] **Step 2: Add `Evaluate` handler method to `handler.go`**

Append to end of `internal/featureflag/interfaces/http/handler.go` (before the closing of the file):

```go
// Evaluate evaluates a single feature flag for the given user attributes.
func (h *Handler) Evaluate(ctx *gin.Context) {
	var req EvaluateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.bc.EvaluateFlag.Handle(ctx.Request.Context(), query.EvaluateQuery{
		Key:       req.Key,
		UserAttrs: req.UserAttrs,
	})
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"key":       result.Key,
		"value":     result.Value,
		"flag_type": result.FlagType,
	})
}

// BatchEvaluate evaluates multiple feature flags for the given user attributes.
func (h *Handler) BatchEvaluate(ctx *gin.Context) {
	var req BatchEvaluateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.bc.BatchEvaluateFlag.Handle(ctx.Request.Context(), query.BatchEvaluateQuery{
		Keys:      req.Keys,
		UserAttrs: req.UserAttrs,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	flags := make(map[string]gin.H, len(result.Flags))
	for k, v := range result.Flags {
		flags[k] = gin.H{"value": v.Value, "flag_type": v.FlagType}
	}

	ctx.JSON(http.StatusOK, gin.H{"flags": flags})
}
```

- [ ] **Step 3: Add routes to `routes.go`**

Add two new routes inside `RegisterRoutes`, after the rule group sub-routes:

```go
	// Evaluate routes
	g.POST("/evaluate", h.Evaluate)
	g.POST("/evaluate/batch", h.BatchEvaluate)
```

- [ ] **Step 4: Verify compilation**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./internal/featureflag/...`
Expected: This will fail because `bc.EvaluateFlag` and `bc.BatchEvaluateFlag` don't exist yet. That's expected — we wire them in the next task.

- [ ] **Step 5: Commit (WIP — won't compile until Task 5)**

```bash
git add internal/featureflag/interfaces/http/request.go internal/featureflag/interfaces/http/handler.go internal/featureflag/interfaces/http/routes.go
git commit -m "feat(featureflag): add evaluate HTTP handler methods and routes (WIP)"
```

---

### Task 5: Wire query handlers into BoundedContext

**Files:**
- Modify: `internal/featureflag/bc.go`

- [ ] **Step 1: Add fields and wiring**

In `internal/featureflag/bc.go`, add the new fields to the `BoundedContext` struct:

```go
	// Queries
	GetFlag           *query.GetHandler
	ListFlags         *query.ListHandler
	EvaluateFlag      *query.EvaluateHandler
	BatchEvaluateFlag *query.BatchEvaluateHandler
```

In the `NewBoundedContext` function return block, add the new handler instantiations:

```go
		EvaluateFlag:      query.NewEvaluateHandler(cachedEval),
		BatchEvaluateFlag: query.NewBatchEvaluateHandler(cachedEval),
```

- [ ] **Step 2: Verify compilation**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go build ./internal/featureflag/...`
Expected: SUCCESS — everything compiles.

- [ ] **Step 3: Run all featureflag tests**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/featureflag/... -v`
Expected: All tests PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/featureflag/bc.go
git commit -m "feat(featureflag): wire EvaluateHandler and BatchEvaluateHandler into BoundedContext"
```

---

### Task 6: Add HTTP handler tests for evaluate endpoints

**Files:**
- Modify: `internal/featureflag/interfaces/http/handler_test.go`

- [ ] **Step 1: Update `setupRouter` to include evaluate handlers**

In `handler_test.go`, the `setupRouter` function creates a `BoundedContext` manually. Update it to include the new fields. First, add a mock evaluator at the top of the file (after existing mocks):

```go
type mockCachedEvaluator struct {
	results map[string]*cache.EvalResult
}

func (m *mockCachedEvaluator) EvaluateFull(_ context.Context, key string, _ map[string]string) *cache.EvalResult {
	if m.results == nil {
		return nil
	}
	return m.results[key]
}
```

Add the `cache` import: `ffcache "gct/internal/featureflag/infrastructure/cache"`

Update the `setupRouter` function signature and body to accept and wire the mock evaluator:

```go
func setupRouter(repo *mockFeatureFlagRepo, rgRepo *mockRuleGroupRepo, readRepo *mockReadRepo) *gin.Engine {
```

Add inside `setupRouter`, after `ListFlags`:

```go
		EvaluateFlag:      query.NewEvaluateHandler(&mockCachedEvaluator{}),
		BatchEvaluateFlag: query.NewBatchEvaluateHandler(&mockCachedEvaluator{}),
```

Also create a variant helper for evaluate tests:

```go
func setupRouterWithEvaluator(evalResults map[string]*cache.EvalResult) *gin.Engine {
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
		GetFlag:           query.NewGetHandler(&mockReadRepo{}),
		ListFlags:         query.NewListHandler(&mockReadRepo{}),
		EvaluateFlag:      query.NewEvaluateHandler(eval),
		BatchEvaluateFlag: query.NewBatchEvaluateHandler(eval),
	}

	r := gin.New()
	h := NewHandler(bc, l)
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)
	return r
}
```

- [ ] **Step 2: Write evaluate endpoint tests**

Append to `handler_test.go`:

```go
func TestHandler_Evaluate_Success(t *testing.T) {
	router := setupRouterWithEvaluator(map[string]*cache.EvalResult{
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
	router := setupRouterWithEvaluator(map[string]*cache.EvalResult{
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
	router := setupRouterWithEvaluator(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/feature-flags/evaluate/batch", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
```

- [ ] **Step 3: Run all handler tests**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/featureflag/interfaces/http/... -v`
Expected: All tests PASS.

- [ ] **Step 4: Run full featureflag test suite**

Run: `cd "/Users/mrb/Desktop/Golang Template/Backend" && go test ./internal/featureflag/... -v`
Expected: All tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/featureflag/interfaces/http/handler_test.go
git commit -m "test(featureflag): add HTTP handler tests for evaluate endpoints"
```
