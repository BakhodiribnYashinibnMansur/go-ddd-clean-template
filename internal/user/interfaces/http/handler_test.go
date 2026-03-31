package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/shared/application"
	shared "gct/internal/shared/domain"
	"gct/internal/user"
	"gct/internal/user/application/command"
	"gct/internal/user/application/query"
	"gct/internal/user/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Mock infrastructure
// ---------------------------------------------------------------------------

type mockUserRepo struct {
	savedUser   *domain.User
	updatedUser *domain.User
	findByIDFn  func(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

func (m *mockUserRepo) Save(_ context.Context, entity *domain.User) error {
	m.savedUser = entity
	return nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) Update(_ context.Context, entity *domain.User) error {
	m.updatedUser = entity
	return nil
}

func (m *mockUserRepo) Delete(_ context.Context, _ uuid.UUID) error { return nil }

func (m *mockUserRepo) List(_ context.Context, _ shared.Pagination) ([]*domain.User, int64, error) {
	return nil, 0, nil
}

func (m *mockUserRepo) FindByPhone(_ context.Context, phone domain.Phone) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) FindByEmail(_ context.Context, email domain.Email) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) FindDefaultRoleID(_ context.Context) (uuid.UUID, error) {
	return uuid.New(), nil
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
	view  *domain.UserView
	views []*domain.UserView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.UserView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ domain.UsersFilter) ([]*domain.UserView, int64, error) {
	return m.views, m.total, nil
}

func (m *mockReadRepo) FindSessionByID(_ context.Context, _ uuid.UUID) (*shared.AuthSession, error) {
	return nil, domain.ErrUserNotFound
}

func (m *mockReadRepo) FindUserForAuth(_ context.Context, _ uuid.UUID) (*shared.AuthUser, error) {
	return nil, domain.ErrUserNotFound
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func setupRouter(bc *user.BoundedContext) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(bc, &mockLogger{})
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)
	return r
}

func newBC(repo *mockUserRepo, readRepo *mockReadRepo) *user.BoundedContext {
	eb := &mockEventBus{}
	l := &mockLogger{}
	return &user.BoundedContext{
		CreateUser:  command.NewCreateUserHandler(repo, eb, l),
		UpdateUser:  command.NewUpdateUserHandler(repo, eb, l),
		DeleteUser:  command.NewDeleteUserHandler(repo, eb, l),
		SignIn:      command.NewSignInHandler(repo, eb, l, command.JWTConfig{}),
		SignUp:      command.NewSignUpHandler(repo, eb, l),
		SignOut:     command.NewSignOutHandler(repo, eb, l),
		ApproveUser: command.NewApproveUserHandler(repo, eb, l),
		ChangeRole:  command.NewChangeRoleHandler(repo, eb, l),
		BulkAction:  command.NewBulkActionHandler(repo, eb, l),
		GetUser:     query.NewGetUserHandler(readRepo),
		ListUsers:   query.NewListUsersHandler(readRepo),
	}
}

// ---------------------------------------------------------------------------
// Tests: POST /users (Create)
// ---------------------------------------------------------------------------

func TestHandler_Create_Success(t *testing.T) {
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	body := CreateUserRequest{
		Phone:    "+998901234567",
		Password: "StrongP@ss123",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	if repo.savedUser == nil {
		t.Fatal("expected user to be saved")
	}
}

func TestHandler_Create_BadRequest(t *testing.T) {
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	// Missing required fields
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: GET /users (List)
// ---------------------------------------------------------------------------

func TestHandler_List_Success(t *testing.T) {
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{
		views: []*domain.UserView{
			{ID: uuid.New(), Phone: "+998901111111", Active: true},
		},
		total: 1,
	}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users?limit=10&offset=0", nil)
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

// ---------------------------------------------------------------------------
// Tests: GET /users/:id (Get)
// ---------------------------------------------------------------------------

func TestHandler_Get_Success(t *testing.T) {
	userID := uuid.New()
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{
		view: &domain.UserView{
			ID:     userID,
			Phone:  "+998901234567",
			Active: true,
		},
	}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users/"+userID.String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Get_InvalidID(t *testing.T) {
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users/not-a-uuid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Get_NotFound(t *testing.T) {
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{} // no view set
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users/"+uuid.New().String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: PATCH /users/:id (Update)
// ---------------------------------------------------------------------------

func TestHandler_Update_Success(t *testing.T) {
	phone, _ := domain.NewPhone("+998901234567")
	pw, _ := domain.NewPasswordFromRaw("StrongP@ss123")
	existingUser := domain.NewUser(phone, pw)

	repo := &mockUserRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
			if id == existingUser.ID() {
				return existingUser, nil
			}
			return nil, domain.ErrUserNotFound
		},
	}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	newEmail := "updated@example.com"
	body := UpdateUserRequest{Email: &newEmail}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/users/"+existingUser.ID().String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Tests: DELETE /users/:id (Delete)
// ---------------------------------------------------------------------------

func TestHandler_Delete_Success(t *testing.T) {
	phone, _ := domain.NewPhone("+998901234567")
	pw, _ := domain.NewPasswordFromRaw("StrongP@ss123")
	existingUser := domain.NewUser(phone, pw)

	repo := &mockUserRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
			if id == existingUser.ID() {
				return existingUser, nil
			}
			return nil, domain.ErrUserNotFound
		},
	}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/users/"+existingUser.ID().String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/users/bad-id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: POST /users/:id/approve (Approve)
// ---------------------------------------------------------------------------

func TestHandler_Approve_Success(t *testing.T) {
	phone, _ := domain.NewPhone("+998901234567")
	pw, _ := domain.NewPasswordFromRaw("StrongP@ss123")
	existingUser := domain.NewUser(phone, pw)

	repo := &mockUserRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
			if id == existingUser.ID() {
				return existingUser, nil
			}
			return nil, domain.ErrUserNotFound
		},
	}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users/"+existingUser.ID().String()+"/approve", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Tests: POST /users/:id/role (ChangeRole)
// ---------------------------------------------------------------------------

func TestHandler_ChangeRole_Success(t *testing.T) {
	phone, _ := domain.NewPhone("+998901234567")
	pw, _ := domain.NewPasswordFromRaw("StrongP@ss123")
	existingUser := domain.NewUser(phone, pw)

	repo := &mockUserRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
			if id == existingUser.ID() {
				return existingUser, nil
			}
			return nil, domain.ErrUserNotFound
		},
	}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	roleID := uuid.New()
	body := ChangeRoleRequest{RoleID: roleID}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users/"+existingUser.ID().String()+"/role", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_ChangeRole_BadRequest(t *testing.T) {
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users/"+uuid.New().String()+"/role", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: POST /users/bulk-action (BulkAction)
// ---------------------------------------------------------------------------

func TestHandler_BulkAction_Success(t *testing.T) {
	phone, _ := domain.NewPhone("+998901234567")
	pw, _ := domain.NewPasswordFromRaw("StrongP@ss123")
	existingUser := domain.NewUser(phone, pw)

	repo := &mockUserRepo{
		findByIDFn: func(_ context.Context, id uuid.UUID) (*domain.User, error) {
			if id == existingUser.ID() {
				return existingUser, nil
			}
			return nil, domain.ErrUserNotFound
		},
	}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	body := BulkActionRequest{
		IDs:    []uuid.UUID{existingUser.ID()},
		Action: "deactivate",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users/bulk-action", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Tests: POST /auth/sign-up (SignUp)
// ---------------------------------------------------------------------------

func TestHandler_SignUp_Success(t *testing.T) {
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	body := SignUpRequest{
		Phone:    "+998901234567",
		Password: "StrongP@ss123",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/sign-up", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_SignUp_BadRequest(t *testing.T) {
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/sign-up", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: POST /auth/sign-in (SignIn)
// ---------------------------------------------------------------------------

func TestHandler_SignIn_BadRequest(t *testing.T) {
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/sign-in", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Tests: POST /auth/sign-out (SignOut)
// ---------------------------------------------------------------------------

func TestHandler_SignOut_BadRequest(t *testing.T) {
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/sign-out", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Additional error path tests
// ---------------------------------------------------------------------------

func TestHandler_Create_InvalidPhone(t *testing.T) {
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	body := CreateUserRequest{Phone: "bad", Password: "StrongP@ss123"}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for invalid phone, got %d", w.Code)
	}
}

func TestHandler_Update_InvalidID(t *testing.T) {
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/users/not-a-uuid", bytes.NewBufferString(`{"email":"a@b.com"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Approve_InvalidID(t *testing.T) {
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users/not-uuid/approve", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_BulkAction_BadRequest(t *testing.T) {
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users/bulk-action", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBufferString(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestHandler_SignIn_InvalidJSON(t *testing.T) {
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/sign-in", bytes.NewBufferString(`not json`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestHandler_ChangeRole_InvalidID(t *testing.T) {
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	roleID := "550e8400-e29b-41d4-a716-446655440000"
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users/bad-id/role",
		bytes.NewBufferString(`{"role_id":"`+roleID+`"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_List_DefaultPagination(t *testing.T) {
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{
		views: []*domain.UserView{},
		total: 0,
	}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	// No query params — should use defaults
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_List_WithFilters(t *testing.T) {
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{
		views: []*domain.UserView{},
		total: 0,
	}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users?phone=+998901234567&email=test@example.com&active=true&limit=5&offset=10", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_ResponseFormat(t *testing.T) {
	userID := uuid.New()
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{
		view: &domain.UserView{
			ID:     userID,
			Phone:  "+998901234567",
			Active: true,
		},
	}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users/"+userID.String(), nil)
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
