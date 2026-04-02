package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/authz"
	"gct/internal/authz/application/command"
	"gct/internal/authz/application/query"
	"gct/internal/authz/domain"
	"gct/internal/shared/application"
	shared "gct/internal/shared/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Mock infrastructure
// ---------------------------------------------------------------------------

type mockRoleRepository struct {
	savedRole   *domain.Role
	updatedRole *domain.Role
	findByIDFn  func(ctx context.Context, id uuid.UUID) (*domain.Role, error)
}

func (m *mockRoleRepository) Save(_ context.Context, role *domain.Role) error {
	m.savedRole = role
	return nil
}

func (m *mockRoleRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, domain.ErrRoleNotFound
}

func (m *mockRoleRepository) Update(_ context.Context, role *domain.Role) error {
	m.updatedRole = role
	return nil
}

func (m *mockRoleRepository) Delete(_ context.Context, _ uuid.UUID) error { return nil }

func (m *mockRoleRepository) List(_ context.Context, _ shared.Pagination) ([]*domain.Role, int64, error) {
	return nil, 0, nil
}

type mockPermissionRepository struct {
	savedPerm *domain.Permission
}

func (m *mockPermissionRepository) Save(_ context.Context, perm *domain.Permission) error {
	m.savedPerm = perm
	return nil
}

func (m *mockPermissionRepository) FindByID(_ context.Context, _ uuid.UUID) (*domain.Permission, error) {
	return nil, domain.ErrPermissionNotFound
}

func (m *mockPermissionRepository) Update(_ context.Context, _ *domain.Permission) error {
	return nil
}

func (m *mockPermissionRepository) Delete(_ context.Context, _ uuid.UUID) error { return nil }

func (m *mockPermissionRepository) List(_ context.Context, _ shared.Pagination) ([]*domain.Permission, int64, error) {
	return nil, 0, nil
}

type mockPolicyRepository struct {
	savedPolicy *domain.Policy
	findByIDFn  func(ctx context.Context, id uuid.UUID) (*domain.Policy, error)
}

func (m *mockPolicyRepository) Save(_ context.Context, policy *domain.Policy) error {
	m.savedPolicy = policy
	return nil
}

func (m *mockPolicyRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Policy, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, domain.ErrPolicyNotFound
}

func (m *mockPolicyRepository) Update(_ context.Context, _ *domain.Policy) error { return nil }

func (m *mockPolicyRepository) Delete(_ context.Context, _ uuid.UUID) error { return nil }

func (m *mockPolicyRepository) List(_ context.Context, _ shared.Pagination) ([]*domain.Policy, int64, error) {
	return nil, 0, nil
}

func (m *mockPolicyRepository) FindByPermissionID(_ context.Context, _ uuid.UUID) ([]*domain.Policy, error) {
	return nil, nil
}

type mockScopeRepository struct{}

func (m *mockScopeRepository) Save(_ context.Context, _ domain.Scope) error  { return nil }
func (m *mockScopeRepository) Delete(_ context.Context, _, _ string) error   { return nil }
func (m *mockScopeRepository) List(_ context.Context, _ shared.Pagination) ([]domain.Scope, int64, error) {
	return nil, 0, nil
}

type mockRolePermissionRepository struct{}

func (m *mockRolePermissionRepository) Assign(_ context.Context, _, _ uuid.UUID) error { return nil }
func (m *mockRolePermissionRepository) Revoke(_ context.Context, _, _ uuid.UUID) error { return nil }

type mockPermissionScopeRepository struct{}

func (m *mockPermissionScopeRepository) Assign(_ context.Context, _ uuid.UUID, _, _ string) error {
	return nil
}

func (m *mockPermissionScopeRepository) Revoke(_ context.Context, _ uuid.UUID, _, _ string) error {
	return nil
}

type mockAuthzReadRepository struct {
	getRoleFn      func(ctx context.Context, id uuid.UUID) (*domain.RoleView, error)
	listRolesFn    func(ctx context.Context, p shared.Pagination) ([]*domain.RoleView, int64, error)
	listPermsFn    func(ctx context.Context, p shared.Pagination) ([]*domain.PermissionView, int64, error)
	listPoliciesFn func(ctx context.Context, p shared.Pagination) ([]*domain.PolicyView, int64, error)
	listScopesFn   func(ctx context.Context, p shared.Pagination) ([]*domain.ScopeView, int64, error)
}

func (m *mockAuthzReadRepository) GetRole(ctx context.Context, id uuid.UUID) (*domain.RoleView, error) {
	if m.getRoleFn != nil {
		return m.getRoleFn(ctx, id)
	}
	return nil, domain.ErrRoleNotFound
}

func (m *mockAuthzReadRepository) ListRoles(ctx context.Context, p shared.Pagination) ([]*domain.RoleView, int64, error) {
	if m.listRolesFn != nil {
		return m.listRolesFn(ctx, p)
	}
	return []*domain.RoleView{}, 0, nil
}

func (m *mockAuthzReadRepository) GetPermission(_ context.Context, _ uuid.UUID) (*domain.PermissionView, error) {
	return nil, domain.ErrPermissionNotFound
}

func (m *mockAuthzReadRepository) ListPermissions(ctx context.Context, p shared.Pagination) ([]*domain.PermissionView, int64, error) {
	if m.listPermsFn != nil {
		return m.listPermsFn(ctx, p)
	}
	return []*domain.PermissionView{}, 0, nil
}

func (m *mockAuthzReadRepository) ListPolicies(ctx context.Context, p shared.Pagination) ([]*domain.PolicyView, int64, error) {
	if m.listPoliciesFn != nil {
		return m.listPoliciesFn(ctx, p)
	}
	return []*domain.PolicyView{}, 0, nil
}

func (m *mockAuthzReadRepository) ListScopes(ctx context.Context, p shared.Pagination) ([]*domain.ScopeView, int64, error) {
	if m.listScopesFn != nil {
		return m.listScopesFn(ctx, p)
	}
	return []*domain.ScopeView{}, 0, nil
}

func (m *mockAuthzReadRepository) CheckAccess(_ context.Context, _ uuid.UUID, _, _ string, _ domain.EvaluationContext) (bool, error) {
	return false, nil
}

func (m *mockAuthzReadRepository) FindPoliciesByPermissionIDs(_ context.Context, _ []uuid.UUID) ([]*domain.Policy, error) {
	return nil, nil
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

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func setupRouter(bc *authz.BoundedContext) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(bc, &mockLogger{})
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)
	return r
}

func newBC(
	roleRepo *mockRoleRepository,
	permRepo *mockPermissionRepository,
	policyRepo *mockPolicyRepository,
	scopeRepo *mockScopeRepository,
	rolePermRepo *mockRolePermissionRepository,
	permScopeRepo *mockPermissionScopeRepository,
	readRepo *mockAuthzReadRepository,
) *authz.BoundedContext {
	eb := &mockEventBus{}
	l := &mockLogger{}
	return &authz.BoundedContext{
		// Commands — Roles
		CreateRole: command.NewCreateRoleHandler(roleRepo, eb, l),
		UpdateRole: command.NewUpdateRoleHandler(roleRepo, eb, l),
		DeleteRole: command.NewDeleteRoleHandler(roleRepo, eb, l),

		// Commands — Permissions
		CreatePermission: command.NewCreatePermissionHandler(permRepo, l),
		DeletePermission: command.NewDeletePermissionHandler(permRepo, l),

		// Commands — Policies
		CreatePolicy: command.NewCreatePolicyHandler(policyRepo, l),
		UpdatePolicy: command.NewUpdatePolicyHandler(policyRepo, l),
		DeletePolicy: command.NewDeletePolicyHandler(policyRepo, l),
		TogglePolicy: command.NewTogglePolicyHandler(policyRepo, l),

		// Commands — Scopes
		CreateScope: command.NewCreateScopeHandler(scopeRepo, l),
		DeleteScope: command.NewDeleteScopeHandler(scopeRepo, l),

		// Commands — Assignments
		AssignPermission: command.NewAssignPermissionHandler(rolePermRepo, eb, l),
		AssignScope:      command.NewAssignScopeHandler(permScopeRepo, l),

		// Queries
		GetRole:         query.NewGetRoleHandler(readRepo, l),
		ListRoles:       query.NewListRolesHandler(readRepo, l),
		ListPermissions: query.NewListPermissionsHandler(readRepo, l),
		ListPolicies:    query.NewListPoliciesHandler(readRepo, l),
		ListScopes:      query.NewListScopesHandler(readRepo, l),
	}
}

func defaultBC() (*authz.BoundedContext, *mockRoleRepository, *mockPolicyRepository) {
	roleRepo := &mockRoleRepository{}
	permRepo := &mockPermissionRepository{}
	policyRepo := &mockPolicyRepository{}
	scopeRepo := &mockScopeRepository{}
	rolePermRepo := &mockRolePermissionRepository{}
	permScopeRepo := &mockPermissionScopeRepository{}
	readRepo := &mockAuthzReadRepository{}
	bc := newBC(roleRepo, permRepo, policyRepo, scopeRepo, rolePermRepo, permScopeRepo, readRepo)
	return bc, roleRepo, policyRepo
}

func defaultBCWithReadRepo(readRepo *mockAuthzReadRepository) *authz.BoundedContext {
	roleRepo := &mockRoleRepository{}
	permRepo := &mockPermissionRepository{}
	policyRepo := &mockPolicyRepository{}
	scopeRepo := &mockScopeRepository{}
	rolePermRepo := &mockRolePermissionRepository{}
	permScopeRepo := &mockPermissionScopeRepository{}
	return newBC(roleRepo, permRepo, policyRepo, scopeRepo, rolePermRepo, permScopeRepo, readRepo)
}

// ---------------------------------------------------------------------------
// Tests: POST /roles (CreateRole)
// ---------------------------------------------------------------------------

func TestHandler_CreateRole_Success(t *testing.T) {
	bc, roleRepo, _ := defaultBC()
	router := setupRouter(bc)

	body := CreateRoleRequest{Name: "admin"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/roles", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if roleRepo.savedRole == nil {
		t.Fatal("expected role to be saved")
	}
	if roleRepo.savedRole.Name() != "admin" {
		t.Errorf("expected name 'admin', got '%s'", roleRepo.savedRole.Name())
	}
}

func TestHandler_CreateRole_WithDescription(t *testing.T) {
	bc, roleRepo, _ := defaultBC()
	router := setupRouter(bc)

	desc := "Administrator role"
	body := CreateRoleRequest{Name: "admin", Description: &desc}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/roles", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if roleRepo.savedRole.Description() == nil || *roleRepo.savedRole.Description() != "Administrator role" {
		t.Error("expected description to be 'Administrator role'")
	}
}

func TestHandler_CreateRole_BadRequest(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	// Missing required "name" field
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/roles", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_CreateRole_InvalidJSON(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/roles", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: GET /roles (ListRoles)
// ---------------------------------------------------------------------------

func TestHandler_ListRoles_Success(t *testing.T) {
	readRepo := &mockAuthzReadRepository{
		listRolesFn: func(_ context.Context, _ shared.Pagination) ([]*domain.RoleView, int64, error) {
			return []*domain.RoleView{
				{ID: uuid.New(), Name: "admin"},
				{ID: uuid.New(), Name: "viewer"},
			}, 2, nil
		},
	}
	bc := defaultBCWithReadRepo(readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/roles?limit=10&offset=0", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	total, ok := resp["total"].(float64)
	if !ok || total != 2 {
		t.Errorf("expected total 2, got %v", resp["total"])
	}
}

func TestHandler_ListRoles_DefaultPagination(t *testing.T) {
	readRepo := &mockAuthzReadRepository{
		listRolesFn: func(_ context.Context, _ shared.Pagination) ([]*domain.RoleView, int64, error) {
			return []*domain.RoleView{}, 0, nil
		},
	}
	bc := defaultBCWithReadRepo(readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/roles", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Tests: GET /roles/:id (GetRole)
// ---------------------------------------------------------------------------

func TestHandler_GetRole_Success(t *testing.T) {
	roleID := uuid.New()
	readRepo := &mockAuthzReadRepository{
		getRoleFn: func(_ context.Context, id uuid.UUID) (*domain.RoleView, error) {
			if id == roleID {
				return &domain.RoleView{ID: roleID, Name: "admin"}, nil
			}
			return nil, domain.ErrRoleNotFound
		},
	}
	bc := defaultBCWithReadRepo(readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/roles/"+roleID.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if _, ok := resp["data"]; !ok {
		t.Error("response should contain 'data' field")
	}
}

func TestHandler_GetRole_InvalidID(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/roles/not-a-uuid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_GetRole_NotFound(t *testing.T) {
	readRepo := &mockAuthzReadRepository{
		getRoleFn: func(_ context.Context, _ uuid.UUID) (*domain.RoleView, error) {
			return nil, domain.ErrRoleNotFound
		},
	}
	bc := defaultBCWithReadRepo(readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/roles/"+uuid.New().String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: PATCH /roles/:id (UpdateRole)
// ---------------------------------------------------------------------------

func TestHandler_UpdateRole_Success(t *testing.T) {
	existingRole := domain.NewRole("old-name")
	roleRepo := &mockRoleRepository{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Role, error) {
			if id == existingRole.ID() {
				return existingRole, nil
			}
			return nil, domain.ErrRoleNotFound
		},
	}
	permRepo := &mockPermissionRepository{}
	policyRepo := &mockPolicyRepository{}
	scopeRepo := &mockScopeRepository{}
	rolePermRepo := &mockRolePermissionRepository{}
	permScopeRepo := &mockPermissionScopeRepository{}
	readRepo := &mockAuthzReadRepository{}
	bc := newBC(roleRepo, permRepo, policyRepo, scopeRepo, rolePermRepo, permScopeRepo, readRepo)
	router := setupRouter(bc)

	newName := "new-name"
	body := UpdateRoleRequest{Name: &newName}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/roles/"+existingRole.ID().String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_UpdateRole_InvalidID(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/roles/bad-id", bytes.NewBufferString(`{"name":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: DELETE /roles/:id (DeleteRole)
// ---------------------------------------------------------------------------

func TestHandler_DeleteRole_Success(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/roles/"+uuid.New().String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_DeleteRole_InvalidID(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/roles/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: POST /permissions (CreatePermission)
// ---------------------------------------------------------------------------

func TestHandler_CreatePermission_Success(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	body := CreatePermissionRequest{Name: "users.read"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/permissions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_CreatePermission_BadRequest(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	// Missing required "name" field
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/permissions", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_CreatePermission_WithParentID(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	parentID := uuid.New()
	body := CreatePermissionRequest{Name: "users.read.self", ParentID: &parentID}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/permissions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Tests: GET /permissions (ListPermissions)
// ---------------------------------------------------------------------------

func TestHandler_ListPermissions_Success(t *testing.T) {
	readRepo := &mockAuthzReadRepository{
		listPermsFn: func(_ context.Context, _ shared.Pagination) ([]*domain.PermissionView, int64, error) {
			return []*domain.PermissionView{
				{ID: uuid.New(), Name: "users.read"},
			}, 1, nil
		},
	}
	bc := defaultBCWithReadRepo(readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/permissions?limit=10&offset=0", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	total, ok := resp["total"].(float64)
	if !ok || total != 1 {
		t.Errorf("expected total 1, got %v", resp["total"])
	}
}

// ---------------------------------------------------------------------------
// Tests: DELETE /permissions/:id (DeletePermission)
// ---------------------------------------------------------------------------

func TestHandler_DeletePermission_Success(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/permissions/"+uuid.New().String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_DeletePermission_InvalidID(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/permissions/not-valid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: POST /policies (CreatePolicy)
// ---------------------------------------------------------------------------

func TestHandler_CreatePolicy_Success(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	permID := uuid.New()
	body := map[string]any{
		"permission_id": permID.String(),
		"effect":        "ALLOW",
		"priority":      10,
		"conditions":    map[string]any{"ip_range": "10.0.0.0/8"},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_CreatePolicy_BadRequest(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	// Missing required "permission_id" and "effect"
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_CreatePolicy_InvalidJSON(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies", bytes.NewBufferString(`not json`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: GET /policies (ListPolicies)
// ---------------------------------------------------------------------------

func TestHandler_ListPolicies_Success(t *testing.T) {
	readRepo := &mockAuthzReadRepository{
		listPoliciesFn: func(_ context.Context, _ shared.Pagination) ([]*domain.PolicyView, int64, error) {
			return []*domain.PolicyView{
				{ID: uuid.New(), PermissionID: uuid.New(), Effect: "ALLOW", Priority: 1, Active: true},
			}, 1, nil
		},
	}
	bc := defaultBCWithReadRepo(readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/policies?limit=10&offset=0", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	total, ok := resp["total"].(float64)
	if !ok || total != 1 {
		t.Errorf("expected total 1, got %v", resp["total"])
	}
}

// ---------------------------------------------------------------------------
// Tests: PATCH /policies/:id (UpdatePolicy)
// ---------------------------------------------------------------------------

func TestHandler_UpdatePolicy_Success(t *testing.T) {
	existingPolicy := domain.NewPolicy(uuid.New(), domain.PolicyAllow)
	policyRepo := &mockPolicyRepository{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Policy, error) {
			if id == existingPolicy.ID() {
				return existingPolicy, nil
			}
			return nil, domain.ErrPolicyNotFound
		},
	}
	roleRepo := &mockRoleRepository{}
	permRepo := &mockPermissionRepository{}
	scopeRepo := &mockScopeRepository{}
	rolePermRepo := &mockRolePermissionRepository{}
	permScopeRepo := &mockPermissionScopeRepository{}
	readRepo := &mockAuthzReadRepository{}
	bc := newBC(roleRepo, permRepo, policyRepo, scopeRepo, rolePermRepo, permScopeRepo, readRepo)
	router := setupRouter(bc)

	newPriority := 99
	body := map[string]any{"priority": newPriority}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/policies/"+existingPolicy.ID().String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_UpdatePolicy_InvalidID(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/policies/bad-id", bytes.NewBufferString(`{"priority":1}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: DELETE /policies/:id (DeletePolicy)
// ---------------------------------------------------------------------------

func TestHandler_DeletePolicy_Success(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/policies/"+uuid.New().String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_DeletePolicy_InvalidID(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/policies/not-uuid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: POST /policies/:id/toggle (TogglePolicy)
// ---------------------------------------------------------------------------

func TestHandler_TogglePolicy_Success(t *testing.T) {
	existingPolicy := domain.NewPolicy(uuid.New(), domain.PolicyAllow)
	policyRepo := &mockPolicyRepository{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.Policy, error) {
			if id == existingPolicy.ID() {
				return existingPolicy, nil
			}
			return nil, domain.ErrPolicyNotFound
		},
	}
	roleRepo := &mockRoleRepository{}
	permRepo := &mockPermissionRepository{}
	scopeRepo := &mockScopeRepository{}
	rolePermRepo := &mockRolePermissionRepository{}
	permScopeRepo := &mockPermissionScopeRepository{}
	readRepo := &mockAuthzReadRepository{}
	bc := newBC(roleRepo, permRepo, policyRepo, scopeRepo, rolePermRepo, permScopeRepo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies/"+existingPolicy.ID().String()+"/toggle", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_TogglePolicy_InvalidID(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/policies/bad-id/toggle", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: POST /scopes (CreateScope)
// ---------------------------------------------------------------------------

func TestHandler_CreateScope_Success(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	body := CreateScopeRequest{Path: "/api/v1/users", Method: "GET"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/scopes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_CreateScope_BadRequest(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	// Missing required "path" and "method"
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/scopes", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: GET /scopes (ListScopes)
// ---------------------------------------------------------------------------

func TestHandler_ListScopes_Success(t *testing.T) {
	readRepo := &mockAuthzReadRepository{
		listScopesFn: func(_ context.Context, _ shared.Pagination) ([]*domain.ScopeView, int64, error) {
			return []*domain.ScopeView{
				{Path: "/api/v1/users", Method: "GET"},
				{Path: "/api/v1/roles", Method: "POST"},
			}, 2, nil
		},
	}
	bc := defaultBCWithReadRepo(readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/scopes?limit=10&offset=0", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	total, ok := resp["total"].(float64)
	if !ok || total != 2 {
		t.Errorf("expected total 2, got %v", resp["total"])
	}
}

// ---------------------------------------------------------------------------
// Tests: DELETE /scopes (DeleteScope)
// ---------------------------------------------------------------------------

func TestHandler_DeleteScope_Success(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	body := DeleteScopeRequest{Path: "/api/v1/users", Method: "GET"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/scopes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_DeleteScope_BadRequest(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	// Missing required fields
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/scopes", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: POST /roles/:id/permissions (AssignPermission)
// ---------------------------------------------------------------------------

func TestHandler_AssignPermission_Success(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	roleID := uuid.New()
	permID := uuid.New()
	body := AssignPermissionRequest{PermissionID: permID}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/roles/"+roleID.String()+"/permissions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_AssignPermission_InvalidRoleID(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	permID := uuid.New()
	body := AssignPermissionRequest{PermissionID: permID}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/roles/bad-id/permissions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_AssignPermission_BadRequest(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	// Missing required "permission_id"
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/roles/"+uuid.New().String()+"/permissions", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: POST /permissions/:id/scopes (AssignScope)
// ---------------------------------------------------------------------------

func TestHandler_AssignScope_Success(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	permID := uuid.New()
	body := AssignScopeRequest{Path: "/api/v1/users", Method: "GET"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/permissions/"+permID.String()+"/scopes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_AssignScope_InvalidPermissionID(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	body := AssignScopeRequest{Path: "/api/v1/users", Method: "GET"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/permissions/bad-id/scopes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_AssignScope_BadRequest(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	// Missing required "path" and "method"
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/permissions/"+uuid.New().String()+"/scopes", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: Response format consistency
// ---------------------------------------------------------------------------

func TestHandler_ListRoles_ResponseFormat(t *testing.T) {
	readRepo := &mockAuthzReadRepository{
		listRolesFn: func(_ context.Context, _ shared.Pagination) ([]*domain.RoleView, int64, error) {
			return []*domain.RoleView{
				{ID: uuid.New(), Name: "admin"},
			}, 1, nil
		},
	}
	bc := defaultBCWithReadRepo(readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/roles", nil)
	router.ServeHTTP(w, req)

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response should be valid JSON: %v", err)
	}
	if _, ok := resp["data"]; !ok {
		t.Error("list response should contain 'data' field")
	}
	if _, ok := resp["total"]; !ok {
		t.Error("list response should contain 'total' field")
	}
}

func TestHandler_CreateRole_ResponseFormat(t *testing.T) {
	bc, _, _ := defaultBC()
	router := setupRouter(bc)

	body := CreateRoleRequest{Name: "viewer"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/roles", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response should be valid JSON: %v", err)
	}
	if success, ok := resp["success"].(bool); !ok || !success {
		t.Error("create response should contain 'success: true'")
	}
}
