package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/kernel/consts"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ---------------------------------------------------------------------------
// getStatusColor
// ---------------------------------------------------------------------------

func TestGetStatusColor(t *testing.T) {
	tests := []struct {
		name   string
		status int
		wantFg string
		wantBg string
	}{
		{"200 OK", 200, logger.ColorBlack, logger.BgBrightGreen},
		{"201 Created", 201, logger.ColorBlack, logger.BgBrightGreen},
		{"299 upper edge 2xx", 299, logger.ColorBlack, logger.BgBrightGreen},
		{"301 Moved Permanently", 301, logger.ColorBlack, logger.BgBrightCyan},
		{"304 Not Modified", 304, logger.ColorBlack, logger.BgBrightCyan},
		{"399 upper edge 3xx", 399, logger.ColorBlack, logger.BgBrightCyan},
		{"400 Bad Request", 400, logger.ColorBlack, logger.BgBrightYellow},
		{"404 Not Found", 404, logger.ColorBlack, logger.BgBrightYellow},
		{"499 upper edge 4xx", 499, logger.ColorBlack, logger.BgBrightYellow},
		{"500 Internal Server Error", 500, logger.ColorBrightWhite, logger.BgRed},
		{"503 Service Unavailable", 503, logger.ColorBrightWhite, logger.BgRed},
		{"100 default", 100, logger.ColorBrightWhite, logger.BgGray},
		{"0 default", 0, logger.ColorBrightWhite, logger.BgGray},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fg, bg := getStatusColor(tt.status)
			if fg != tt.wantFg {
				t.Errorf("getStatusColor(%d) fg = %q, want %q", tt.status, fg, tt.wantFg)
			}
			if bg != tt.wantBg {
				t.Errorf("getStatusColor(%d) bg = %q, want %q", tt.status, bg, tt.wantBg)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// getMethodStyle
// ---------------------------------------------------------------------------

func TestGetMethodStyle(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		wantFg    string
		wantBg    string
		wantLabel string
	}{
		{"GET", "GET", logger.ColorBlack, logger.BgCyan, " GET "},
		{"POST", "POST", logger.ColorBlack, logger.BgGreen, " POST "},
		{"PUT", "PUT", logger.ColorBlack, logger.BgOrange, " PUT "},
		{"DELETE", "DELETE", logger.ColorBrightWhite, logger.BgRed, " DEL "},
		{"PATCH", "PATCH", logger.ColorBlack, logger.BgMagenta, " PATCH "},
		{"HEAD", "HEAD", logger.ColorBlack, logger.BgWhite, " HEAD "},
		{"OPTIONS", "OPTIONS", logger.ColorBrightWhite, logger.BgPurple, " OPT "},
		{"unknown method", "TRACE", logger.ColorBlack, logger.BgGray, " ??? "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fg, bg, label := getMethodStyle(tt.method)
			if fg != tt.wantFg {
				t.Errorf("getMethodStyle(%q) fg = %q, want %q", tt.method, fg, tt.wantFg)
			}
			if bg != tt.wantBg {
				t.Errorf("getMethodStyle(%q) bg = %q, want %q", tt.method, bg, tt.wantBg)
			}
			if label != tt.wantLabel {
				t.Errorf("getMethodStyle(%q) label = %q, want %q", tt.method, label, tt.wantLabel)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Logger middleware
// ---------------------------------------------------------------------------

// newLoggerTestRouter builds a Gin engine with the Logger middleware and a
// single GET /test handler that replies with the given status code.
func newLoggerTestRouter(status int) *gin.Engine {
	r := gin.New()
	r.Use(Logger(&mockLog{}))
	r.GET("/test", func(c *gin.Context) {
		c.Status(status)
	})
	return r
}

func TestLoggerSetsXRequestIDWhenNotProvided(t *testing.T) {
	r := newLoggerTestRouter(http.StatusOK)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	rid := w.Header().Get(consts.HeaderXRequestID)
	if rid == "" {
		t.Fatal("expected X-Request-ID header to be set, got empty string")
	}
	// UUID v4 format: 8-4-4-4-12 hex characters = 36 chars total.
	if len(rid) != 36 {
		t.Errorf("expected UUID-length X-Request-ID (36 chars), got %d chars: %q", len(rid), rid)
	}
}

func TestLoggerPreservesExistingXRequestID(t *testing.T) {
	r := newLoggerTestRouter(http.StatusOK)

	customID := "my-custom-request-id-12345"
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(consts.HeaderXRequestID, customID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	got := w.Header().Get(consts.HeaderXRequestID)
	if got != customID {
		t.Errorf("expected X-Request-ID %q, got %q", customID, got)
	}
}

func TestLoggerSetsRequestIDInContext(t *testing.T) {
	r := gin.New()
	r.Use(Logger(&mockLog{}))

	var ctxRequestID string
	r.GET("/test", func(c *gin.Context) {
		ctxRequestID = c.GetString(consts.CtxKeyRequestID)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if ctxRequestID == "" {
		t.Fatal("expected request_id to be set in gin context, got empty string")
	}

	headerID := w.Header().Get(consts.HeaderXRequestID)
	if ctxRequestID != headerID {
		t.Errorf("context request_id %q does not match header X-Request-ID %q", ctxRequestID, headerID)
	}
}

func TestLoggerStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		wantStatus int
	}{
		{"200 OK", http.StatusOK, http.StatusOK},
		{"404 Not Found", http.StatusNotFound, http.StatusNotFound},
		{"500 Internal Server Error", http.StatusInternalServerError, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newLoggerTestRouter(tt.status)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}
