package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/context/iam/generic/user"
	"gct/internal/context/iam/generic/user/application/command"
	"gct/internal/context/iam/generic/user/application/query"
	userentity "gct/internal/context/iam/generic/user/domain/entity"
	"gct/internal/kernel/application"
	shared "gct/internal/kernel/domain"

	"gct/internal/kernel/outbox"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Mock infrastructure
// ---------------------------------------------------------------------------

type mockUserRepo struct {
	savedUser   *userentity.User
	updatedUser *userentity.User
	findByIDFn  func(ctx context.Context, id userentity.UserID) (*userentity.User, error)
}

func (m *mockUserRepo) Save(_ context.Context, entity *userentity.User) error {
	m.savedUser = entity
	return nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id userentity.UserID) (*userentity.User, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, userentity.ErrUserNotFound
}

func (m *mockUserRepo) Update(_ context.Context, entity *userentity.User) error {
	m.updatedUser = entity
	return nil
}

func (m *mockUserRepo) Delete(_ context.Context, _ userentity.UserID) error { return nil }

func (m *mockUserRepo) List(_ context.Context, _ shared.Pagination) ([]*userentity.User, int64, error) {
	return nil, 0, nil
}

func (m *mockUserRepo) FindByPhone(_ context.Context, phone userentity.Phone) (*userentity.User, error) {
	return nil, userentity.ErrUserNotFound
}

func (m *mockUserRepo) FindByEmail(_ context.Context, email userentity.Email) (*userentity.User, error) {
	return nil, userentity.ErrUserNotFound
}

func (m *mockUserRepo) FindDefaultRoleID(_ context.Context) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockUserRepo) ActiveSessionCount(_ context.Context, _ userentity.UserID) (int, error) {
	return 0, nil
}

func (m *mockUserRepo) RevokeOldestActiveSession(_ context.Context, _ userentity.UserID) (userentity.SessionID, error) {
	return userentity.NilSessionID, nil
}

func (m *mockUserRepo) RevokeSessionsByIntegration(_ context.Context, _ userentity.UserID, _ string) (int, error) {
	return 0, nil
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

type mockReadRepo struct {
	view  *userentity.UserView
	views []*userentity.UserView
	total int64
}

func (m *mockReadRepo) FindByID(_ context.Context, id userentity.UserID) (*userentity.UserView, error) {
	if m.view != nil && m.view.ID == id {
		return m.view, nil
	}
	return nil, userentity.ErrUserNotFound
}

func (m *mockReadRepo) List(_ context.Context, _ userentity.UsersFilter) ([]*userentity.UserView, int64, error) {
	return m.views, m.total, nil
}

func (m *mockReadRepo) FindSessionByID(_ context.Context, _ userentity.SessionID) (*shared.AuthSession, error) {
	return nil, userentity.ErrUserNotFound
}

func (m *mockReadRepo) FindUserForAuth(_ context.Context, _ userentity.UserID) (*shared.AuthUser, error) {
	return nil, userentity.ErrUserNotFound
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
		CreateUser:  command.NewCreateUserHandler(repo, outbox.NewEventCommitter(nil, nil, eb, l), l),
		UpdateUser:  command.NewUpdateUserHandler(repo, outbox.NewEventCommitter(nil, nil, eb, l), l),
		DeleteUser:  command.NewDeleteUserHandler(repo, outbox.NewEventCommitter(nil, nil, eb, l), l),
		SignIn:      command.NewSignInHandler(repo, outbox.NewEventCommitter(nil, nil, eb, l), l, command.JWTConfig{}),
		SignUp:      command.NewSignUpHandler(repo, eb, l),
		SignOut:     command.NewSignOutHandler(repo, eb, l),
		ApproveUser: command.NewApproveUserHandler(repo, eb, l),
		ChangeRole:  command.NewChangeRoleHandler(repo, eb, l),
		BulkAction:  command.NewBulkActionHandler(repo, eb, l),
		GetUser:     query.NewGetUserHandler(readRepo, l),
		ListUsers:   query.NewListUsersHandler(readRepo, l),
	}
}

// ---------------------------------------------------------------------------
// Tests: POST /users (Create)
// ---------------------------------------------------------------------------

func TestHandler_Create_Success(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{
		views: []*userentity.UserView{
			{ID: userentity.NewUserID(), Phone: "+998901111111", Active: true},
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
	t.Parallel()

	userID := userentity.NewUserID()
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{
		view: &userentity.UserView{
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
	t.Parallel()

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
	t.Parallel()

	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{} // no view set
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users/"+uuid.New().String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Tests: PATCH /users/:id (Update)
// ---------------------------------------------------------------------------

func TestHandler_Update_Success(t *testing.T) {
	t.Parallel()

	phone, _ := userentity.NewPhone("+998901234567")
	pw, _ := userentity.NewPasswordFromRaw("StrongP@ss123")
	existingUser, _ := userentity.NewUser(phone, pw)

	repo := &mockUserRepo{
		findByIDFn: func(_ context.Context, id userentity.UserID) (*userentity.User, error) {
			if id == existingUser.TypedID() {
				return existingUser, nil
			}
			return nil, userentity.ErrUserNotFound
		},
	}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	newEmail := "updated@example.com"
	body := UpdateUserRequest{Email: &newEmail}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/users/"+existingUser.TypedID().String(), bytes.NewBuffer(jsonBody))
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
	t.Parallel()

	phone, _ := userentity.NewPhone("+998901234567")
	pw, _ := userentity.NewPasswordFromRaw("StrongP@ss123")
	existingUser, _ := userentity.NewUser(phone, pw)

	repo := &mockUserRepo{
		findByIDFn: func(_ context.Context, id userentity.UserID) (*userentity.User, error) {
			if id == existingUser.TypedID() {
				return existingUser, nil
			}
			return nil, userentity.ErrUserNotFound
		},
	}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/users/"+existingUser.TypedID().String(), nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	phone, _ := userentity.NewPhone("+998901234567")
	pw, _ := userentity.NewPasswordFromRaw("StrongP@ss123")
	existingUser, _ := userentity.NewUser(phone, pw)

	repo := &mockUserRepo{
		findByIDFn: func(_ context.Context, id userentity.UserID) (*userentity.User, error) {
			if id == existingUser.TypedID() {
				return existingUser, nil
			}
			return nil, userentity.ErrUserNotFound
		},
	}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users/"+existingUser.TypedID().String()+"/approve", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Tests: POST /users/:id/role (ChangeRole)
// ---------------------------------------------------------------------------

func TestHandler_ChangeRole_Success(t *testing.T) {
	t.Parallel()

	phone, _ := userentity.NewPhone("+998901234567")
	pw, _ := userentity.NewPasswordFromRaw("StrongP@ss123")
	existingUser, _ := userentity.NewUser(phone, pw)

	repo := &mockUserRepo{
		findByIDFn: func(_ context.Context, id userentity.UserID) (*userentity.User, error) {
			if id == existingUser.TypedID() {
				return existingUser, nil
			}
			return nil, userentity.ErrUserNotFound
		},
	}
	readRepo := &mockReadRepo{}
	bc := newBC(repo, readRepo)
	router := setupRouter(bc)

	roleID := uuid.New()
	body := ChangeRoleRequest{RoleID: roleID}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users/"+existingUser.TypedID().String()+"/role", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_ChangeRole_BadRequest(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	phone, _ := userentity.NewPhone("+998901234567")
	pw, _ := userentity.NewPasswordFromRaw("StrongP@ss123")
	existingUser, _ := userentity.NewUser(phone, pw)

	repo := &mockUserRepo{
		findByIDFn: func(_ context.Context, id userentity.UserID) (*userentity.User, error) {
			if id == existingUser.TypedID() {
				return existingUser, nil
			}
			return nil, userentity.ErrUserNotFound
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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid phone, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Update_InvalidID(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{
		views: []*userentity.UserView{},
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
	t.Parallel()

	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{
		views: []*userentity.UserView{},
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
	t.Parallel()

	userID := userentity.NewUserID()
	repo := &mockUserRepo{}
	readRepo := &mockReadRepo{
		view: &userentity.UserView{
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

// --- Additional error-path, parsing, and pagination tests ---

func TestHandler_GetUser_InvalidUUID(t *testing.T) {
	bc := newBC(&mockUserRepo{}, &mockReadRepo{})
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users/not-a-uuid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid UUID, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_DeleteUser_InvalidUUID(t *testing.T) {
	bc := newBC(&mockUserRepo{}, &mockReadRepo{})
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/users/not-a-uuid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid UUID, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_ListUsers_DefaultPagination(t *testing.T) {
	readRepo := &mockReadRepo{
		views: []*userentity.UserView{},
		total: 0,
	}
	bc := newBC(&mockUserRepo{}, readRepo)
	router := setupRouter(bc)

	// No query params — should use default pagination and succeed
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for default pagination, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_CreateUser_EmptyBody(t *testing.T) {
	bc := newBC(&mockUserRepo{}, &mockReadRepo{})
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBufferString(``))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty body, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_CreateUser_InvalidJSON(t *testing.T) {
	bc := newBC(&mockUserRepo{}, &mockReadRepo{})
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBufferString(`{invalid json`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid JSON, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_UpdateUser_InvalidUUID(t *testing.T) {
	bc := newBC(&mockUserRepo{}, &mockReadRepo{})
	router := setupRouter(bc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/api/v1/users/not-a-uuid", bytes.NewBufferString(`{"email":"a@b.com"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid UUID, got %d: %s", w.Code, w.Body.String())
	}
}
