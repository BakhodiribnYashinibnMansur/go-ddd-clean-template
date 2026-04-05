package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	access "gct/internal/context/iam/authz/application/query"
	"gct/internal/context/iam/authz/domain"
	"gct/internal/contract/ports"
	shared "gct/internal/platform/domain"
	"gct/internal/platform/domain/consts"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Mock infrastructure
// ---------------------------------------------------------------------------

type mockLogger struct{}

func (m *mockLogger) Debug(args ...any)                                                {}
func (m *mockLogger) Debugf(template string, args ...any)                              {}
func (m *mockLogger) Debugw(msg string, keysAndValues ...any)                          {}
func (m *mockLogger) Info(args ...any)                                                 {}
func (m *mockLogger) Infof(template string, args ...any)                               {}
func (m *mockLogger) Infow(msg string, keysAndValues ...any)                           {}
func (m *mockLogger) Warn(args ...any)                                                 {}
func (m *mockLogger) Warnf(template string, args ...any)                               {}
func (m *mockLogger) Warnw(msg string, keysAndValues ...any)                           {}
func (m *mockLogger) Error(args ...any)                                                {}
func (m *mockLogger) Errorf(template string, args ...any)                              {}
func (m *mockLogger) Errorw(msg string, keysAndValues ...any)                          {}
func (m *mockLogger) Fatal(args ...any)                                                {}
func (m *mockLogger) Fatalf(template string, args ...any)                              {}
func (m *mockLogger) Fatalw(msg string, keysAndValues ...any)                          {}
func (m *mockLogger) Debugc(_ context.Context, _ string, _ ...any)                     {}
func (m *mockLogger) Infoc(_ context.Context, _ string, _ ...any)                      {}
func (m *mockLogger) Warnc(_ context.Context, _ string, _ ...any)                      {}
func (m *mockLogger) Errorc(_ context.Context, _ string, _ ...any)                     {}
func (m *mockLogger) Fatalc(_ context.Context, _ string, _ ...any)                     {}

type mockAuthzReadRepository struct {
	checkAccessFn func(ctx context.Context, roleID uuid.UUID, path, method string, evalCtx domain.EvaluationContext) (bool, error)
}

func (m *mockAuthzReadRepository) GetRole(context.Context, uuid.UUID) (*domain.RoleView, error) {
	return nil, nil
}
func (m *mockAuthzReadRepository) ListRoles(context.Context, shared.Pagination) ([]*domain.RoleView, int64, error) {
	return nil, 0, nil
}
func (m *mockAuthzReadRepository) GetPermission(context.Context, uuid.UUID) (*domain.PermissionView, error) {
	return nil, nil
}
func (m *mockAuthzReadRepository) ListPermissions(context.Context, shared.Pagination) ([]*domain.PermissionView, int64, error) {
	return nil, 0, nil
}
func (m *mockAuthzReadRepository) ListPolicies(context.Context, shared.Pagination) ([]*domain.PolicyView, int64, error) {
	return nil, 0, nil
}
func (m *mockAuthzReadRepository) ListScopes(context.Context, shared.Pagination) ([]*domain.ScopeView, int64, error) {
	return nil, 0, nil
}
func (m *mockAuthzReadRepository) CheckAccess(ctx context.Context, roleID uuid.UUID, path, method string, evalCtx domain.EvaluationContext) (bool, error) {
	if m.checkAccessFn != nil {
		return m.checkAccessFn(ctx, roleID, path, method, evalCtx)
	}
	return false, nil
}

func (m *mockAuthzReadRepository) FindPoliciesByPermissionIDs(_ context.Context, _ []uuid.UUID) ([]*domain.Policy, error) {
	return nil, nil
}

// fakeAuthUserLookup is a test double satisfying ports.AuthUserLookup so the
// authz middleware can be exercised without importing the user BC.
type fakeAuthUserLookup struct {
	findFn func(ctx context.Context, userID uuid.UUID) (*shared.AuthUser, error)
}

func (f *fakeAuthUserLookup) FindForAuth(ctx context.Context, userID uuid.UUID) (*shared.AuthUser, error) {
	if f.findFn != nil {
		return f.findFn(ctx, userID)
	}
	return nil, errors.New("user not found")
}

var _ ports.AuthUserLookup = (*fakeAuthUserLookup)(nil)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func setupMiddleware(
	checkAccessFn func(ctx context.Context, roleID uuid.UUID, path, method string, evalCtx domain.EvaluationContext) (bool, error),
	findUserFn func(ctx context.Context, userID uuid.UUID) (*shared.AuthUser, error),
) *AuthzMiddleware {
	l := &mockLogger{}
	readRepo := &mockAuthzReadRepository{checkAccessFn: checkAccessFn}
	userLookup := &fakeAuthUserLookup{findFn: findUserFn}

	checkAccessHandler := access.NewCheckAccessHandler(readRepo, l)

	return NewAuthzMiddleware(checkAccessHandler, userLookup, l)
}

func performRequest(mw *AuthzMiddleware, method, path string, sessionSetup func(*gin.Context)) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		if sessionSetup != nil {
			sessionSetup(c)
		}
		c.Next()
	})
	r.Use(mw.Authz)

	r.Handle(method, path, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	r.ServeHTTP(w, req)
	return w
}

func validSession() *shared.AuthSession {
	return &shared.AuthSession{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		DeviceID:     uuid.New(),
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Revoked:      false,
		LastActivity: time.Now(),
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestAuthzMiddleware_NoSession(t *testing.T) {
	mw := setupMiddleware(nil, nil)

	w := performRequest(mw, "GET", "/api/v1/users", nil)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthzMiddleware_InvalidSessionType(t *testing.T) {
	mw := setupMiddleware(nil, nil)

	w := performRequest(mw, "GET", "/api/v1/users", func(c *gin.Context) {
		c.Set(consts.CtxSession, "not-a-session")
	})

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestAuthzMiddleware_UserNotFound(t *testing.T) {
	session := validSession()

	mw := setupMiddleware(nil, func(_ context.Context, _ uuid.UUID) (*shared.AuthUser, error) {
		return nil, errors.New("user not found")
	})

	w := performRequest(mw, "GET", "/api/v1/users", func(c *gin.Context) {
		c.Set(consts.CtxSession, session)
	})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthzMiddleware_UserHasNoRole(t *testing.T) {
	session := validSession()

	mw := setupMiddleware(nil, func(_ context.Context, userID uuid.UUID) (*shared.AuthUser, error) {
		return &shared.AuthUser{
			ID:     userID,
			RoleID: nil,
			Active: true,
		}, nil
	})

	w := performRequest(mw, "GET", "/api/v1/users", func(c *gin.Context) {
		c.Set(consts.CtxSession, session)
	})

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestAuthzMiddleware_AccessAllowed(t *testing.T) {
	session := validSession()
	roleID := uuid.New()

	mw := setupMiddleware(
		func(_ context.Context, _ uuid.UUID, _ string, _ string, _ domain.EvaluationContext) (bool, error) {
			return true, nil
		},
		func(_ context.Context, userID uuid.UUID) (*shared.AuthUser, error) {
			return &shared.AuthUser{
				ID:     userID,
				RoleID: &roleID,
				Active: true,
			}, nil
		},
	)

	w := performRequest(mw, "GET", "/api/v1/users", func(c *gin.Context) {
		c.Set(consts.CtxSession, session)
	})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestAuthzMiddleware_AccessDenied(t *testing.T) {
	session := validSession()
	roleID := uuid.New()

	mw := setupMiddleware(
		func(_ context.Context, _ uuid.UUID, _ string, _ string, _ domain.EvaluationContext) (bool, error) {
			return false, nil
		},
		func(_ context.Context, userID uuid.UUID) (*shared.AuthUser, error) {
			return &shared.AuthUser{
				ID:     userID,
				RoleID: &roleID,
				Active: true,
			}, nil
		},
	)

	w := performRequest(mw, "GET", "/api/v1/users", func(c *gin.Context) {
		c.Set(consts.CtxSession, session)
	})

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestAuthzMiddleware_CheckAccessError(t *testing.T) {
	session := validSession()
	roleID := uuid.New()

	mw := setupMiddleware(
		func(_ context.Context, _ uuid.UUID, _ string, _ string, _ domain.EvaluationContext) (bool, error) {
			return false, errors.New("db connection error")
		},
		func(_ context.Context, userID uuid.UUID) (*shared.AuthUser, error) {
			return &shared.AuthUser{
				ID:     userID,
				RoleID: &roleID,
				Active: true,
			}, nil
		},
	)

	w := performRequest(mw, "GET", "/api/v1/users", func(c *gin.Context) {
		c.Set(consts.CtxSession, session)
	})

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestAuthzMiddleware_CorrectRoleIDPassedToCheckAccess(t *testing.T) {
	session := validSession()
	roleID := uuid.New()
	var capturedRoleID uuid.UUID
	var capturedPath string
	var capturedMethod string

	mw := setupMiddleware(
		func(_ context.Context, rID uuid.UUID, path string, method string, _ domain.EvaluationContext) (bool, error) {
			capturedRoleID = rID
			capturedPath = path
			capturedMethod = method
			return true, nil
		},
		func(_ context.Context, userID uuid.UUID) (*shared.AuthUser, error) {
			return &shared.AuthUser{
				ID:     userID,
				RoleID: &roleID,
				Active: true,
			}, nil
		},
	)

	performRequest(mw, "POST", "/api/v1/users", func(c *gin.Context) {
		c.Set(consts.CtxSession, session)
	})

	if capturedRoleID != roleID {
		t.Errorf("expected roleID %s, got %s", roleID, capturedRoleID)
	}
	if capturedPath != "/api/v1/users" {
		t.Errorf("expected path /api/v1/users, got %s", capturedPath)
	}
	if capturedMethod != "POST" {
		t.Errorf("expected method POST, got %s", capturedMethod)
	}
}

func TestAuthzMiddleware_CorrectUserIDPassedToFindUser(t *testing.T) {
	session := validSession()
	var capturedUserID uuid.UUID

	mw := setupMiddleware(
		nil,
		func(_ context.Context, userID uuid.UUID) (*shared.AuthUser, error) {
			capturedUserID = userID
			return nil, errors.New("not found")
		},
	)

	performRequest(mw, "GET", "/api/v1/users", func(c *gin.Context) {
		c.Set(consts.CtxSession, session)
	})

	if capturedUserID != session.UserID {
		t.Errorf("expected userID %s, got %s", session.UserID, capturedUserID)
	}
}

func TestAuthzMiddleware_DifferentHTTPMethods(t *testing.T) {
	session := validSession()
	roleID := uuid.New()

	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			var capturedMethod string

			mw := setupMiddleware(
				func(_ context.Context, _ uuid.UUID, _ string, m string, _ domain.EvaluationContext) (bool, error) {
					capturedMethod = m
					return true, nil
				},
				func(_ context.Context, userID uuid.UUID) (*shared.AuthUser, error) {
					return &shared.AuthUser{
						ID:     userID,
						RoleID: &roleID,
						Active: true,
					}, nil
				},
			)

			w := performRequest(mw, method, "/api/v1/users", func(c *gin.Context) {
				c.Set(consts.CtxSession, session)
			})

			if w.Code != http.StatusOK {
				t.Errorf("expected 200 for %s, got %d", method, w.Code)
			}
			if capturedMethod != method {
				t.Errorf("expected captured method %s, got %s", method, capturedMethod)
			}
		})
	}
}

func TestAuthzMiddleware_AbortsPipelineOnDenied(t *testing.T) {
	session := validSession()
	roleID := uuid.New()
	handlerCalled := false

	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set(consts.CtxSession, session)
		c.Next()
	})

	mw := setupMiddleware(
		func(_ context.Context, _ uuid.UUID, _ string, _ string, _ domain.EvaluationContext) (bool, error) {
			return false, nil
		},
		func(_ context.Context, userID uuid.UUID) (*shared.AuthUser, error) {
			return &shared.AuthUser{
				ID:     userID,
				RoleID: &roleID,
				Active: true,
			}, nil
		},
	)

	r.Use(mw.Authz)
	r.GET("/api/v1/secret", func(c *gin.Context) {
		handlerCalled = true
		c.JSON(http.StatusOK, nil)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/secret", nil)
	r.ServeHTTP(w, req)

	if handlerCalled {
		t.Error("expected handler NOT to be called when access is denied")
	}
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestAuthzMiddleware_AllowsPipelineOnGranted(t *testing.T) {
	session := validSession()
	roleID := uuid.New()
	handlerCalled := false

	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set(consts.CtxSession, session)
		c.Next()
	})

	mw := setupMiddleware(
		func(_ context.Context, _ uuid.UUID, _ string, _ string, _ domain.EvaluationContext) (bool, error) {
			return true, nil
		},
		func(_ context.Context, userID uuid.UUID) (*shared.AuthUser, error) {
			return &shared.AuthUser{
				ID:     userID,
				RoleID: &roleID,
				Active: true,
			}, nil
		},
	)

	r.Use(mw.Authz)
	r.GET("/api/v1/open", func(c *gin.Context) {
		handlerCalled = true
		c.JSON(http.StatusOK, nil)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/open", nil)
	r.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("expected handler to be called when access is granted")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
