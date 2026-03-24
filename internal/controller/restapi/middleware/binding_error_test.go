package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type testBindingPayload struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

func TestBindingErrorMiddleware_NoErrorPassesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(BindingErrorMiddleware())
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBindingErrorMiddleware_BindingErrorReturnsJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(BindingErrorMiddleware())
	r.POST("/test", func(c *gin.Context) {
		var payload testBindingPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			// Gin adds binding errors to c.Errors when ShouldBind fails
			// but BindingErrorMiddleware catches errors added via c.Error()
			// We need to manually add the error for the middleware to catch
			_ = c.Error(err).SetType(gin.ErrorTypeBind)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Send invalid JSON (missing required fields)
	body := `{"name": ""}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var respBody map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &respBody)
	assert.NoError(t, err)
	assert.Equal(t, "error", respBody["status"])
}

func TestBindingErrorMiddleware_ValidPayloadPasses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(BindingErrorMiddleware())
	r.POST("/test", func(c *gin.Context) {
		var payload testBindingPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			_ = c.Error(err).SetType(gin.ErrorTypeBind)
			return
		}
		c.JSON(http.StatusOK, gin.H{"name": payload.Name})
	})

	body := `{"name": "John", "email": "john@example.com"}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBindingErrorMiddleware_NonBindingErrorIgnored(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(BindingErrorMiddleware())
	r.POST("/test", func(c *gin.Context) {
		// Add a non-binding error (private type)
		_ = c.Error(gin.Error{Err: assert.AnError, Type: gin.ErrorTypePrivate}.Err)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	r.ServeHTTP(w, req)

	// The private error should not trigger a 400 response
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBindingErrorMiddleware_EmptyBodyBindingError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(BindingErrorMiddleware())
	r.POST("/test", func(c *gin.Context) {
		var payload testBindingPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			_ = c.Error(err).SetType(gin.ErrorTypeBind)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Send completely empty body
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
