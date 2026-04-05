package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/internal/kernel/consts"

	"github.com/gin-gonic/gin"
)

func newProdConfig() *config.Config {
	return &config.Config{
		App:      config.App{Environment: "production"},
		Security: config.Security{FetchMetadata: true},
	}
}

func newDevConfig() *config.Config {
	return &config.Config{
		App:      config.App{Environment: "development"},
		Security: config.Security{FetchMetadata: true},
	}
}

func TestFetchMetadata_NonProdBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(FetchMetadata(newDevConfig()))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 in non-prod, got %d", w.Code)
	}
}

func TestFetchMetadata_DisabledBypass(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	cfg := &config.Config{
		App:      config.App{Environment: "production"},
		Security: config.Security{FetchMetadata: false},
	}

	r.Use(FetchMetadata(cfg))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 when disabled, got %d", w.Code)
	}
}

func TestFetchMetadata_MissingSecFetchSite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(FetchMetadata(newProdConfig()))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	// No Sec-Fetch-Site header
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for missing Sec-Fetch-Site, got %d", w.Code)
	}
}

func TestFetchMetadata_SameOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(FetchMetadata(newProdConfig()))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set(consts.HeaderSecFetchSite, consts.HeaderValueSameOrigin)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for same-origin, got %d", w.Code)
	}
}

func TestFetchMetadata_SameSite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(FetchMetadata(newProdConfig()))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set(consts.HeaderSecFetchSite, consts.HeaderValueSameSite)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for same-site, got %d", w.Code)
	}
}

func TestFetchMetadata_TopLevelNavigation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(FetchMetadata(newProdConfig()))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set(consts.HeaderSecFetchSite, "cross-site")
	req.Header.Set(consts.HeaderSecFetchMode, consts.HeaderValueNavigate)
	req.Header.Set(consts.HeaderSecFetchDest, consts.HeaderValueDocument)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for top-level navigation, got %d", w.Code)
	}
}

func TestFetchMetadata_TopLevelNavPOSTBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(FetchMetadata(newProdConfig()))
	r.POST("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("POST", "/test", nil)
	req.Header.Set(consts.HeaderSecFetchSite, "cross-site")
	req.Header.Set(consts.HeaderSecFetchMode, consts.HeaderValueNavigate)
	req.Header.Set(consts.HeaderSecFetchDest, consts.HeaderValueDocument)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for POST cross-site navigation, got %d", w.Code)
	}
}

func TestFetchMetadata_CrossSiteBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	r.Use(FetchMetadata(newProdConfig()))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set(consts.HeaderSecFetchSite, "cross-site")
	req.Header.Set(consts.HeaderSecFetchMode, "cors")
	req.Header.Set(consts.HeaderSecFetchDest, "script")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for cross-site non-navigation, got %d", w.Code)
	}
}
