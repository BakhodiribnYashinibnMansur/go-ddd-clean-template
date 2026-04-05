package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestMethodNotAllowedHandler_DirectInvocation tests the handler function directly.
// MethodNotAllowedHandler is designed to be passed to gin's r.NoMethod(),
// which is called by gin when a route exists but the HTTP method is wrong.
// Note: The handler passes ErrHandlerMethodNotAllowed through RespondWithError,
// which uses the error's internal code mapping to determine the status code.
func TestMethodNotAllowedHandler_DirectInvocation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	handler := MethodNotAllowedHandler()
	r.GET("/method-not-allowed", handler)

	c.Request, _ = http.NewRequest("GET", "/method-not-allowed", nil)
	r.ServeHTTP(w, c.Request)

	// The handler calls RespondWithError with the error code.
	// The response code depends on the error code mapping in the response package.
	if w.Code == http.StatusOK {
		t.Error("expected non-200 status for method not allowed handler")
	}
}

func TestMethodNotAllowedHandler_ReturnsJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	handler := MethodNotAllowedHandler()
	r.GET("/method-not-allowed", handler)

	c.Request, _ = http.NewRequest("GET", "/method-not-allowed", nil)
	r.ServeHTTP(w, c.Request)

	ct := w.Header().Get("Content-Type")
	if ct == "" {
		t.Fatal("expected Content-Type header to be set")
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON response body, got error: %v", err)
	}

	status, ok := body["status"].(string)
	if !ok || status != "error" {
		t.Errorf("expected status 'error', got %v", body["status"])
	}

	// Verify the error contains an error object
	errObj, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatal("expected error object in response")
	}

	if errObj["code"] == nil || errObj["code"] == "" {
		t.Error("expected error code to be set")
	}
}

func TestMethodNotAllowedHandler_AllowedMethodPasses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	// Register route with allowed method; MethodNotAllowedHandler is only for NoMethod
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for allowed method, got %d", w.Code)
	}
}
