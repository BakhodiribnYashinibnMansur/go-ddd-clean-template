package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBindingErrorMiddleware_NoErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(BindingErrorMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestBindingErrorMiddleware_BindError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(BindingErrorMiddleware())
	r.GET("/test", func(c *gin.Context) {
		// Simulate a binding error by adding one to the context
		_ = c.Error(gin.Error{
			Err:  http.ErrBodyNotAllowed,
			Type: gin.ErrorTypeBind,
		}.Err).SetType(gin.ErrorTypeBind)
		// Do not write a response so the middleware can handle it
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for bind error, got %d", w.Code)
	}
}

func TestBindingErrorMiddleware_NonBindError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(BindingErrorMiddleware())
	r.GET("/test", func(c *gin.Context) {
		_ = c.Error(http.ErrBodyNotAllowed).SetType(gin.ErrorTypePrivate)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for non-bind error, got %d", w.Code)
	}
}
