package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/config"
	"gct/internal/shared/domain/consts"
	"gct/internal/domain"
	"gct/internal/usecase/authz"
	"gct/internal/usecase/authz/access"
	"gct/internal/usecase/authz/permission"
	"gct/internal/usecase/authz/policy"
	"gct/internal/usecase/authz/relation"
	"gct/internal/usecase/authz/role"
	"gct/internal/usecase/authz/scope"
	"gct/internal/usecase/integration"
	"gct/internal/usecase/user/client"
	"gct/internal/usecase/user/session"
	"gct/internal/shared/infrastructure/security/jwt"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ========== Mock Implementations ==========

type mockUserUC struct {
	mock.Mock
}

var _ client.UseCaseI = (*mockUserUC)(nil)

func (m *mockUserUC) Create(ctx context.Context, in *domain.User) error {
	return m.Called(ctx, in).Error(0)
}
func (m *mockUserUC) Get(ctx context.Context, in *domain.UserFilter) (*domain.User, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *mockUserUC) Gets(ctx context.Context, in *domain.UsersFilter) ([]*domain.User, int, error) {
	args := m.Called(ctx, in)
	return args.Get(0).([]*domain.User), args.Int(1), args.Error(2)
}
func (m *mockUserUC) Update(ctx context.Context, in *domain.User) error {
	return m.Called(ctx, in).Error(0)
}
func (m *mockUserUC) Delete(ctx context.Context, in *domain.UserFilter) error {
	return m.Called(ctx, in).Error(0)
}
func (m *mockUserUC) SignIn(ctx context.Context, in *domain.SignInIn) (*domain.SignInOut, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*domain.SignInOut), args.Error(1)
}
func (m *mockUserUC) SignUp(ctx context.Context, in *domain.SignUpIn) (*domain.SignInOut, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*domain.SignInOut), args.Error(1)
}
func (m *mockUserUC) SignOut(ctx context.Context, in *domain.SignOutIn) error {
	return m.Called(ctx, in).Error(0)
}
func (m *mockUserUC) RotateSession(ctx context.Context, in *domain.RefreshIn) (*domain.SignInOut, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*domain.SignInOut), args.Error(1)
}
func (m *mockUserUC) GetByPhone(ctx context.Context, in *domain.UserFilter) (*domain.User, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *mockUserUC) ActivateUser(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *mockUserUC) SetStatus(ctx context.Context, id uuid.UUID, active bool) error {
	return m.Called(ctx, id, active).Error(0)
}
func (m *mockUserUC) BulkAction(ctx context.Context, req domain.BulkActionRequest) error {
	return m.Called(ctx, req).Error(0)
}
func (m *mockUserUC) Approve(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserUC) ChangeRole(ctx context.Context, id, roleStr string) error {
	return m.Called(ctx, id, roleStr).Error(0)
}

type mockSessionUC struct {
	mock.Mock
}

var _ session.UseCaseI = (*mockSessionUC)(nil)

func (m *mockSessionUC) Create(ctx context.Context, in *domain.Session) (*domain.Session, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*domain.Session), args.Error(1)
}
func (m *mockSessionUC) Get(ctx context.Context, in *domain.SessionFilter) (*domain.Session, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Session), args.Error(1)
}
func (m *mockSessionUC) Gets(ctx context.Context, in *domain.SessionsFilter) ([]*domain.Session, int, error) {
	args := m.Called(ctx, in)
	return args.Get(0).([]*domain.Session), args.Int(1), args.Error(2)
}
func (m *mockSessionUC) Update(ctx context.Context, in *domain.Session) error {
	return m.Called(ctx, in).Error(0)
}
func (m *mockSessionUC) UpdateActivity(ctx context.Context, in *domain.SessionFilter) error {
	return m.Called(ctx, in).Error(0)
}
func (m *mockSessionUC) Revoke(ctx context.Context, in *domain.SessionFilter) error {
	return m.Called(ctx, in).Error(0)
}
func (m *mockSessionUC) Delete(ctx context.Context, in *domain.SessionFilter) error {
	return m.Called(ctx, in).Error(0)
}

type mockRoleUC struct {
	mock.Mock
}

var _ role.UseCaseI = (*mockRoleUC)(nil)

func (m *mockRoleUC) Create(ctx context.Context, in *domain.Role) error {
	return m.Called(ctx, in).Error(0)
}
func (m *mockRoleUC) Get(ctx context.Context, in *domain.RoleFilter) (*domain.Role, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}
func (m *mockRoleUC) Gets(ctx context.Context, in *domain.RolesFilter) ([]*domain.Role, int, error) {
	args := m.Called(ctx, in)
	return args.Get(0).([]*domain.Role), args.Int(1), args.Error(2)
}
func (m *mockRoleUC) Update(ctx context.Context, in *domain.Role) error {
	return m.Called(ctx, in).Error(0)
}
func (m *mockRoleUC) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockRoleUC) Assign(ctx context.Context, userID, roleID uuid.UUID) error {
	return m.Called(ctx, userID, roleID).Error(0)
}
func (m *mockRoleUC) AddPermission(ctx context.Context, roleID, permID uuid.UUID) error {
	return m.Called(ctx, roleID, permID).Error(0)
}
func (m *mockRoleUC) RemovePermission(ctx context.Context, roleID, permID uuid.UUID) error {
	return m.Called(ctx, roleID, permID).Error(0)
}

type mockAuthzUC struct {
	accessUC    access.UseCaseI
	roleUC      role.UseCaseI
}

var _ authz.UseCaseI = (*mockAuthzUC)(nil)

func (m *mockAuthzUC) Access() access.UseCaseI       { return m.accessUC }
func (m *mockAuthzUC) Role() role.UseCaseI           { return m.roleUC }
func (m *mockAuthzUC) Permission() permission.UseCaseI { return nil }
func (m *mockAuthzUC) Policy() policy.UseCaseI       { return nil }
func (m *mockAuthzUC) Relation() relation.UseCaseI   { return nil }
func (m *mockAuthzUC) Scope() scope.UseCaseI         { return nil }

// ========== Test Helpers ==========

func generateTestRSAKeys(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)
	return privateKey, &privateKey.PublicKey
}

func newTestAuthMiddleware(t *testing.T, userUC *mockUserUC, sessionUC *mockSessionUC, authzUC *mockAuthzUC) (*AuthMiddleware, *rsa.PrivateKey) {
	t.Helper()
	privateKey, publicKey := generateTestRSAKeys(t)
	l := logger.New("debug")

	m := &AuthMiddleware{
		userUC:    userUC,
		sessionuc: sessionUC,
		authzUC:   authzUC,
		cfg: &config.Config{
			JWT: config.JWT{
				Issuer: "test-issuer",
			},
		},
		l:      l,
		pubKey: publicKey,
	}
	return m, privateKey
}

func generateValidAccessToken(t *testing.T, privateKey *rsa.PrivateKey, userID, sessionID, issuer string, ttl time.Duration) string {
	t.Helper()
	token, err := jwt.GenerateAccessToken(userID, sessionID, issuer, "", privateKey, ttl)
	assert.NoError(t, err)
	return token
}

// ========== AuthClientAccess Tests ==========

func TestAuthClientAccess_MissingTokenReturns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userUC := new(mockUserUC)
	sessionUC := new(mockSessionUC)
	m, _ := newTestAuthMiddleware(t, userUC, sessionUC, nil)

	r := gin.New()
	r.GET("/test", m.AuthClientAccess, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthClientAccess_InvalidTokenReturns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userUC := new(mockUserUC)
	sessionUC := new(mockSessionUC)
	m, _ := newTestAuthMiddleware(t, userUC, sessionUC, nil)

	r := gin.New()
	r.GET("/test", m.AuthClientAccess, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-jwt-token")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthClientAccess_MalformedAuthHeaderReturns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userUC := new(mockUserUC)
	sessionUC := new(mockSessionUC)
	m, _ := newTestAuthMiddleware(t, userUC, sessionUC, nil)

	r := gin.New()
	r.GET("/test", m.AuthClientAccess, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "NotBearer some-token")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthClientAccess_ExpiredTokenReturns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userUC := new(mockUserUC)
	sessionUC := new(mockSessionUC)
	m, privateKey := newTestAuthMiddleware(t, userUC, sessionUC, nil)

	sessionID := uuid.New()
	userID := uuid.New()

	// Generate an expired token (TTL of -1 hour)
	token := generateValidAccessToken(t, privateKey, userID.String(), sessionID.String(), "test-issuer", -1*time.Hour)

	r := gin.New()
	r.GET("/test", m.AuthClientAccess, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthClientAccess_WrongIssuerReturns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userUC := new(mockUserUC)
	sessionUC := new(mockSessionUC)
	m, privateKey := newTestAuthMiddleware(t, userUC, sessionUC, nil)

	sessionID := uuid.New()
	userID := uuid.New()

	// Generate token with wrong issuer
	token := generateValidAccessToken(t, privateKey, userID.String(), sessionID.String(), "wrong-issuer", 1*time.Hour)

	r := gin.New()
	r.GET("/test", m.AuthClientAccess, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthClientAccess_RevokedSessionReturns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userUC := new(mockUserUC)
	sessionUC := new(mockSessionUC)
	m, privateKey := newTestAuthMiddleware(t, userUC, sessionUC, nil)

	sessionID := uuid.New()
	userID := uuid.New()
	token := generateValidAccessToken(t, privateKey, userID.String(), sessionID.String(), "test-issuer", 1*time.Hour)

	// Session is revoked
	sessionUC.On("Get", mock.Anything, mock.Anything).Return(&domain.Session{
		ID:        sessionID,
		UserID:    userID,
		Revoked:   true,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}, nil)

	r := gin.New()
	r.GET("/test", m.AuthClientAccess, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthClientAccess_ExpiredSessionReturns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userUC := new(mockUserUC)
	sessionUC := new(mockSessionUC)
	m, privateKey := newTestAuthMiddleware(t, userUC, sessionUC, nil)

	sessionID := uuid.New()
	userID := uuid.New()
	token := generateValidAccessToken(t, privateKey, userID.String(), sessionID.String(), "test-issuer", 1*time.Hour)

	// Session is expired
	sessionUC.On("Get", mock.Anything, mock.Anything).Return(&domain.Session{
		ID:        sessionID,
		UserID:    userID,
		Revoked:   false,
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}, nil)

	r := gin.New()
	r.GET("/test", m.AuthClientAccess, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthClientAccess_ValidTokenPasses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userUC := new(mockUserUC)
	sessionUC := new(mockSessionUC)
	m, privateKey := newTestAuthMiddleware(t, userUC, sessionUC, nil)

	sessionID := uuid.New()
	userID := uuid.New()
	token := generateValidAccessToken(t, privateKey, userID.String(), sessionID.String(), "test-issuer", 1*time.Hour)

	// Session is valid
	sessionUC.On("Get", mock.Anything, mock.Anything).Return(&domain.Session{
		ID:        sessionID,
		UserID:    userID,
		Revoked:   false,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}, nil)

	r := gin.New()
	r.GET("/test", m.AuthClientAccess, func(c *gin.Context) {
		// Verify context was properly injected
		ctxUserID, exists := c.Get(consts.CtxUserID)
		assert.True(t, exists)
		assert.Equal(t, userID.String(), ctxUserID)

		ctxSessionData, exists := c.Get(consts.CtxSession)
		assert.True(t, exists)
		assert.NotNil(t, ctxSessionData)

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	sessionUC.AssertExpectations(t)
}

func TestAuthClientAccess_ValidTokenViaCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userUC := new(mockUserUC)
	sessionUC := new(mockSessionUC)
	m, privateKey := newTestAuthMiddleware(t, userUC, sessionUC, nil)

	sessionID := uuid.New()
	userID := uuid.New()
	token := generateValidAccessToken(t, privateKey, userID.String(), sessionID.String(), "test-issuer", 1*time.Hour)

	sessionUC.On("Get", mock.Anything, mock.Anything).Return(&domain.Session{
		ID:        sessionID,
		UserID:    userID,
		Revoked:   false,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}, nil)

	r := gin.New()
	r.GET("/test", m.AuthClientAccess, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// Set token via cookie instead of Authorization header
	req.AddCookie(&http.Cookie{Name: consts.CookieAccessToken, Value: token})
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	sessionUC.AssertExpectations(t)
}

func TestAuthClientAccess_WrongRSAKeyReturns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userUC := new(mockUserUC)
	sessionUC := new(mockSessionUC)
	m, _ := newTestAuthMiddleware(t, userUC, sessionUC, nil)

	// Generate token with a DIFFERENT private key
	otherPrivateKey, _ := generateTestRSAKeys(t)
	sessionID := uuid.New()
	userID := uuid.New()
	token := generateValidAccessToken(t, otherPrivateKey, userID.String(), sessionID.String(), "test-issuer", 1*time.Hour)

	r := gin.New()
	r.GET("/test", m.AuthClientAccess, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ========== AuthApiKey Tests ==========

func TestAuthApiKey_MissingKeyReturns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userUC := new(mockUserUC)
	sessionUC := new(mockSessionUC)
	m, _ := newTestAuthMiddleware(t, userUC, sessionUC, nil)

	r := gin.New()
	r.GET("/test", m.AuthApiKey, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthApiKey_DeprecatedAlwaysReturnsForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userUC := new(mockUserUC)
	sessionUC := new(mockSessionUC)
	m, _ := newTestAuthMiddleware(t, userUC, sessionUC, nil)

	r := gin.New()
	r.GET("/test", m.AuthApiKey, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(consts.HeaderXAPIKey, "some-api-key")
	r.ServeHTTP(w, req)

	// AuthApiKey is deprecated and always returns 403
	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== Authz Tests ==========

func TestAuthz_MissingSessionReturns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userUC := new(mockUserUC)
	sessionUC := new(mockSessionUC)
	m, _ := newTestAuthMiddleware(t, userUC, sessionUC, nil)

	r := gin.New()
	r.GET("/test", m.Authz, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthz_SuperAdminBypassesCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userUC := new(mockUserUC)
	sessionUC := new(mockSessionUC)
	mockRole := new(mockRoleUC)

	authzMock := &mockAuthzUC{roleUC: mockRole}
	m, _ := newTestAuthMiddleware(t, userUC, sessionUC, authzMock)

	sessionID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()

	userUC.On("Get", mock.Anything, mock.Anything).Return(&domain.User{
		ID:     userID,
		RoleID: &roleID,
	}, nil)

	mockRole.On("Get", mock.Anything, mock.Anything).Return(&domain.Role{
		ID:   roleID,
		Name: "super_admin",
	}, nil)

	r := gin.New()
	// Inject session into context before Authz
	r.Use(func(c *gin.Context) {
		c.Set(consts.CtxSession, &domain.Session{
			ID:        sessionID,
			UserID:    userID,
			Revoked:   false,
			ExpiresAt: time.Now().Add(1 * time.Hour),
		})
		c.Next()
	})
	r.GET("/test", m.Authz, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	userUC.AssertExpectations(t)
	mockRole.AssertExpectations(t)
}

func TestAuthz_UserWithNoRoleReturns403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userUC := new(mockUserUC)
	sessionUC := new(mockSessionUC)
	mockRole := new(mockRoleUC)

	authzMock := &mockAuthzUC{roleUC: mockRole}
	m, _ := newTestAuthMiddleware(t, userUC, sessionUC, authzMock)

	sessionID := uuid.New()
	userID := uuid.New()

	userUC.On("Get", mock.Anything, mock.Anything).Return(&domain.User{
		ID:     userID,
		RoleID: nil, // No role
	}, nil)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(consts.CtxSession, &domain.Session{
			ID:        sessionID,
			UserID:    userID,
			Revoked:   false,
			ExpiresAt: time.Now().Add(1 * time.Hour),
		})
		c.Next()
	})
	r.GET("/test", m.Authz, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ========== AuthAdmin Tests ==========

func TestAuthAdmin_MissingTokenReturns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userUC := new(mockUserUC)
	sessionUC := new(mockSessionUC)
	m, _ := newTestAuthMiddleware(t, userUC, sessionUC, nil)

	r := gin.New()
	r.GET("/test", m.AuthAdmin, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthAdmin_NonAdminRoleReturns403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userUC := new(mockUserUC)
	sessionUC := new(mockSessionUC)
	mockRole := new(mockRoleUC)

	authzMock := &mockAuthzUC{roleUC: mockRole}
	m, privateKey := newTestAuthMiddleware(t, userUC, sessionUC, authzMock)

	sessionID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()
	token := generateValidAccessToken(t, privateKey, userID.String(), sessionID.String(), "test-issuer", 1*time.Hour)

	sessionUC.On("Get", mock.Anything, mock.Anything).Return(&domain.Session{
		ID:        sessionID,
		UserID:    userID,
		Revoked:   false,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}, nil)

	userUC.On("Get", mock.Anything, mock.Anything).Return(&domain.User{
		ID:     userID,
		RoleID: &roleID,
	}, nil)

	mockRole.On("Get", mock.Anything, mock.Anything).Return(&domain.Role{
		ID:   roleID,
		Name: "user", // Not an admin role
	}, nil)

	r := gin.New()
	r.GET("/test", m.AuthAdmin, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAuthAdmin_AdminRolePasses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userUC := new(mockUserUC)
	sessionUC := new(mockSessionUC)
	mockRole := new(mockRoleUC)

	authzMock := &mockAuthzUC{roleUC: mockRole}
	m, privateKey := newTestAuthMiddleware(t, userUC, sessionUC, authzMock)

	sessionID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()
	token := generateValidAccessToken(t, privateKey, userID.String(), sessionID.String(), "test-issuer", 1*time.Hour)

	sessionUC.On("Get", mock.Anything, mock.Anything).Return(&domain.Session{
		ID:        sessionID,
		UserID:    userID,
		Revoked:   false,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}, nil)

	userUC.On("Get", mock.Anything, mock.Anything).Return(&domain.User{
		ID:     userID,
		RoleID: &roleID,
	}, nil)

	mockRole.On("Get", mock.Anything, mock.Anything).Return(&domain.Role{
		ID:   roleID,
		Name: "admin",
	}, nil)

	r := gin.New()
	r.GET("/test", m.AuthAdmin, func(c *gin.Context) {
		// Verify admin context values
		isAdmin, exists := c.Get(consts.CtxIsAdmin)
		assert.True(t, exists)
		assert.Equal(t, true, isAdmin)

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	sessionUC.AssertExpectations(t)
	userUC.AssertExpectations(t)
	mockRole.AssertExpectations(t)
}

func TestAuthAdmin_SuperAdminRolePasses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userUC := new(mockUserUC)
	sessionUC := new(mockSessionUC)
	mockRole := new(mockRoleUC)

	authzMock := &mockAuthzUC{roleUC: mockRole}
	m, privateKey := newTestAuthMiddleware(t, userUC, sessionUC, authzMock)

	sessionID := uuid.New()
	userID := uuid.New()
	roleID := uuid.New()
	token := generateValidAccessToken(t, privateKey, userID.String(), sessionID.String(), "test-issuer", 1*time.Hour)

	sessionUC.On("Get", mock.Anything, mock.Anything).Return(&domain.Session{
		ID:        sessionID,
		UserID:    userID,
		Revoked:   false,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}, nil)

	userUC.On("Get", mock.Anything, mock.Anything).Return(&domain.User{
		ID:     userID,
		RoleID: &roleID,
	}, nil)

	mockRole.On("Get", mock.Anything, mock.Anything).Return(&domain.Role{
		ID:   roleID,
		Name: "super_admin", // Contains "admin"
	}, nil)

	r := gin.New()
	r.GET("/test", m.AuthAdmin, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ========== Unused interface stubs ==========

type mockIntegrationUC struct{}

var _ integration.UseCaseI = (*mockIntegrationUC)(nil)

func (m *mockIntegrationUC) CreateIntegration(context.Context, domain.CreateIntegrationRequest) (*domain.Integration, error) {
	return nil, nil
}
func (m *mockIntegrationUC) GetIntegration(context.Context, uuid.UUID) (*domain.IntegrationWithKeys, error) {
	return nil, nil
}
func (m *mockIntegrationUC) ListIntegrations(context.Context, domain.IntegrationFilter) ([]domain.Integration, int64, error) {
	return nil, 0, nil
}
func (m *mockIntegrationUC) UpdateIntegration(context.Context, uuid.UUID, domain.UpdateIntegrationRequest) (*domain.Integration, error) {
	return nil, nil
}
func (m *mockIntegrationUC) DeleteIntegration(context.Context, uuid.UUID) error { return nil }
func (m *mockIntegrationUC) ToggleIntegration(context.Context, uuid.UUID) (*domain.Integration, error) {
	return nil, nil
}
func (m *mockIntegrationUC) CreateAPIKey(context.Context, domain.CreateAPIKeyRequest) (*domain.APIKey, string, error) {
	return nil, "", nil
}
func (m *mockIntegrationUC) GetAPIKey(context.Context, uuid.UUID) (*domain.APIKey, error) {
	return nil, nil
}
func (m *mockIntegrationUC) ListAPIKeys(context.Context, uuid.UUID) ([]domain.APIKey, error) {
	return nil, nil
}
func (m *mockIntegrationUC) ValidateAPIKey(context.Context, string) (*domain.APIKey, error) {
	return nil, nil
}
func (m *mockIntegrationUC) RevokeAPIKey(context.Context, uuid.UUID) error { return nil }
func (m *mockIntegrationUC) DeleteAPIKey(context.Context, uuid.UUID) error  { return nil }
func (m *mockIntegrationUC) InitCache(context.Context) error                { return nil }
func (m *mockIntegrationUC) InvalidateCache(context.Context, string) error  { return nil }
