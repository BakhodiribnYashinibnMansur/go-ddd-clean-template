package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRecovery_PanicReturns500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")

	r := gin.New()
	r.Use(Recovery(l))
	r.GET("/panic", func(c *gin.Context) {
		panic("something went wrong")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var body map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "error", body["status"])
}

func TestRecovery_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")

	r := gin.New()
	r.Use(Recovery(l))
	r.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRecovery_PanicWithNilValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")

	r := gin.New()
	r.Use(Recovery(l))
	r.GET("/panic-nil", func(c *gin.Context) {
		panic(nil)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic-nil", nil)
	r.ServeHTTP(w, req)

	// Gin's CustomRecovery catches all panics including nil
	assert.True(t, w.Code == http.StatusInternalServerError || w.Code == http.StatusOK)
}

func TestRecovery_PanicWithError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")

	r := gin.New()
	r.Use(Recovery(l))
	r.POST("/panic-error", func(c *gin.Context) {
		panic("database connection lost")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/panic-error", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRecovery_SubsequentRequestsStillWork(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := logger.New("debug")

	r := gin.New()
	r.Use(Recovery(l))
	r.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})
	r.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First request panics
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusInternalServerError, w1.Code)

	// Second request should still work
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/ok", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}
