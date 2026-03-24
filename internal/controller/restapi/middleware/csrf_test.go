package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/shared/domain/consts"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ---------- Simple CSRF Middleware (Double-Submit Cookie) ----------

func TestCSRFMiddleware_GetRequestSkipsCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")

	r := gin.New()
	r.Use(Middleware(l, consts.CookieCsrfToken))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFMiddleware_HeadRequestSkipsCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")

	r := gin.New()
	r.Use(Middleware(l, consts.CookieCsrfToken))
	r.HEAD("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFMiddleware_OptionsRequestSkipsCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")

	r := gin.New()
	r.Use(Middleware(l, consts.CookieCsrfToken))
	r.OPTIONS("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFMiddleware_PostMissingCookieReturns403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")

	r := gin.New()
	r.Use(Middleware(l, consts.CookieCsrfToken))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRFMiddleware_PostMissingHeaderReturns403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")

	r := gin.New()
	r.Use(Middleware(l, consts.CookieCsrfToken))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	// Set cookie but no header
	req.AddCookie(&http.Cookie{Name: consts.CookieCsrfToken, Value: "token123"})
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRFMiddleware_PostMismatchedTokensReturns403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")

	r := gin.New()
	r.Use(Middleware(l, consts.CookieCsrfToken))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.AddCookie(&http.Cookie{Name: consts.CookieCsrfToken, Value: "token_cookie"})
	req.Header.Set(consts.HeaderXCSRFToken, "token_header_different")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRFMiddleware_PostMatchingTokensPass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")

	r := gin.New()
	r.Use(Middleware(l, consts.CookieCsrfToken))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	token := "valid-csrf-token-12345"
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.AddCookie(&http.Cookie{Name: consts.CookieCsrfToken, Value: token})
	req.Header.Set(consts.HeaderXCSRFToken, token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFMiddleware_PutMatchingTokensPass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")

	r := gin.New()
	r.Use(Middleware(l, consts.CookieCsrfToken))
	r.PUT("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	token := "valid-csrf-token-put"
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/test", nil)
	req.AddCookie(&http.Cookie{Name: consts.CookieCsrfToken, Value: token})
	req.Header.Set(consts.HeaderXCSRFToken, token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRFMiddleware_DeleteMissingCookieReturns403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")

	r := gin.New()
	r.Use(Middleware(l, consts.CookieCsrfToken))
	r.DELETE("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/test", nil)
	req.Header.Set(consts.HeaderXCSRFToken, "some-token")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRFMiddleware_PatchMatchingTokensPass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")

	r := gin.New()
	r.Use(Middleware(l, consts.CookieCsrfToken))
	r.PATCH("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	token := "valid-csrf-token-patch"
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/test", nil)
	req.AddCookie(&http.Cookie{Name: consts.CookieCsrfToken, Value: token})
	req.Header.Set(consts.HeaderXCSRFToken, token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ---------- Hybrid CSRF Middleware ----------

func TestHybridCSRFMiddleware_BearerAuthSkipsCSRF(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")

	r := gin.New()
	r.Use(HybridMiddleware(l, consts.CookieCsrfToken))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Authorization", "Bearer some-jwt-token")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHybridCSRFMiddleware_NoBearerEnforcesCSRF(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")

	r := gin.New()
	r.Use(HybridMiddleware(l, consts.CookieCsrfToken))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// No Authorization header and no CSRF tokens -> should fail
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHybridCSRFMiddleware_NoBearerWithValidCSRFPasses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")

	r := gin.New()
	r.Use(HybridMiddleware(l, consts.CookieCsrfToken))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	token := "valid-csrf-token"
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.AddCookie(&http.Cookie{Name: consts.CookieCsrfToken, Value: token})
	req.Header.Set(consts.HeaderXCSRFToken, token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHybridCSRFMiddleware_GetAlwaysPasses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")

	r := gin.New()
	r.Use(HybridMiddleware(l, consts.CookieCsrfToken))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ---------- Error response body verification ----------

func TestCSRFMiddleware_ErrorResponseContainsMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")

	r := gin.New()
	r.Use(Middleware(l, consts.CookieCsrfToken))
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "error", body["status"])
}
