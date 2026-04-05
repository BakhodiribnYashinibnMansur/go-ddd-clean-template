package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/platform/domain/consts"
	apperrors "gct/internal/platform/infrastructure/errors"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func errorTestGinContext(method, path string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, nil)
	return c, w
}

// ---------------------------------------------------------------------------
// MapToHTTPStatus: repo codes
// ---------------------------------------------------------------------------

func TestMapToHTTPStatus_RepoNotFound(t *testing.T) {
	tests := []struct {
		code string
		want int
	}{
		{apperrors.CodeRepoNotFound, 404},
		{apperrors.ErrRepoNotFound, 404},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := MapToHTTPStatus(tt.code)
			if got != tt.want {
				t.Errorf("MapToHTTPStatus(%q) = %d, want %d", tt.code, got, tt.want)
			}
		})
	}
}

func TestMapToHTTPStatus_RepoAlreadyExists(t *testing.T) {
	tests := []struct {
		code string
		want int
	}{
		{apperrors.CodeRepoAlreadyExists, 409},
		{apperrors.ErrRepoAlreadyExists, 409},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := MapToHTTPStatus(tt.code)
			if got != tt.want {
				t.Errorf("MapToHTTPStatus(%q) = %d, want %d", tt.code, got, tt.want)
			}
		})
	}
}

func TestMapToHTTPStatus_RepoTimeout(t *testing.T) {
	tests := []struct {
		code string
		want int
	}{
		{apperrors.CodeRepoTimeout, 504},
		{apperrors.ErrRepoTimeout, 504},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := MapToHTTPStatus(tt.code)
			if got != tt.want {
				t.Errorf("MapToHTTPStatus(%q) = %d, want %d", tt.code, got, tt.want)
			}
		})
	}
}

func TestMapToHTTPStatus_UserNotFound(t *testing.T) {
	tests := []struct {
		code string
		want int
	}{
		{apperrors.CodeUserNotFound, 404},
		{apperrors.ErrUserNotFound, 404},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := MapToHTTPStatus(tt.code)
			if got != tt.want {
				t.Errorf("MapToHTTPStatus(%q) = %d, want %d", tt.code, got, tt.want)
			}
		})
	}
}

func TestMapToHTTPStatus_SessionNotFound(t *testing.T) {
	tests := []struct {
		code string
		want int
	}{
		{apperrors.CodeSessionNotFound, 404},
		{apperrors.ErrSessionNotFound, 404},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := MapToHTTPStatus(tt.code)
			if got != tt.want {
				t.Errorf("MapToHTTPStatus(%q) = %d, want %d", tt.code, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// MapToHTTPStatus: service codes
// ---------------------------------------------------------------------------

func TestMapToHTTPStatus_ServiceValidation(t *testing.T) {
	codes := []string{
		apperrors.CodeServiceInvalidInput, apperrors.ErrServiceInvalidInput,
		apperrors.CodeServiceValidation, apperrors.ErrServiceValidation,
		apperrors.ErrBadRequest, apperrors.CodeBadRequest,
		apperrors.ErrInvalidInput, apperrors.CodeInvalidInput,
		apperrors.ErrValidation, apperrors.CodeValidation,
	}
	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			got := MapToHTTPStatus(code)
			if got != 400 {
				t.Errorf("MapToHTTPStatus(%q) = %d, want 400", code, got)
			}
		})
	}
}

func TestMapToHTTPStatus_ServiceNotFound(t *testing.T) {
	codes := []string{
		apperrors.CodeServiceNotFound, apperrors.ErrServiceNotFound,
		apperrors.ErrNotFound, apperrors.CodeNotFound,
	}
	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			got := MapToHTTPStatus(code)
			if got != 404 {
				t.Errorf("MapToHTTPStatus(%q) = %d, want 404", code, got)
			}
		})
	}
}

func TestMapToHTTPStatus_ServiceConflict(t *testing.T) {
	codes := []string{
		apperrors.CodeServiceAlreadyExists, apperrors.ErrServiceAlreadyExists,
		apperrors.CodeServiceConflict, apperrors.ErrServiceConflict,
	}
	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			got := MapToHTTPStatus(code)
			if got != 409 {
				t.Errorf("MapToHTTPStatus(%q) = %d, want 409", code, got)
			}
		})
	}
}

func TestMapToHTTPStatus_ServiceUnauthorized(t *testing.T) {
	codes := []string{
		apperrors.CodeServiceUnauthorized, apperrors.ErrServiceUnauthorized,
	}
	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			got := MapToHTTPStatus(code)
			if got != 401 {
				t.Errorf("MapToHTTPStatus(%q) = %d, want 401", code, got)
			}
		})
	}
}

func TestMapToHTTPStatus_ServiceForbidden(t *testing.T) {
	codes := []string{
		apperrors.CodeServiceForbidden, apperrors.ErrServiceForbidden,
	}
	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			got := MapToHTTPStatus(code)
			if got != 403 {
				t.Errorf("MapToHTTPStatus(%q) = %d, want 403", code, got)
			}
		})
	}
}

func TestMapToHTTPStatus_ServiceBusinessRule(t *testing.T) {
	codes := []string{
		apperrors.CodeServiceBusinessRule, apperrors.ErrServiceBusinessRule,
	}
	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			got := MapToHTTPStatus(code)
			if got != 422 {
				t.Errorf("MapToHTTPStatus(%q) = %d, want 422", code, got)
			}
		})
	}
}

func TestMapToHTTPStatus_ServiceDependency(t *testing.T) {
	codes := []string{
		apperrors.CodeServiceDependency, apperrors.ErrServiceDependency,
	}
	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			got := MapToHTTPStatus(code)
			if got != 502 {
				t.Errorf("MapToHTTPStatus(%q) = %d, want 502", code, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// MapToHTTPStatus: handler codes (numeric)
// ---------------------------------------------------------------------------

func TestMapToHTTPStatus_HandlerNumericCodes(t *testing.T) {
	tests := []struct {
		code string
		want int
	}{
		{"4000", 400},
		{"4001", 401},
		{"4003", 403},
		{"4004", 404},
		{"4009", 409},
		{"4029", 429},
	}
	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := MapToHTTPStatus(tt.code)
			if got != tt.want {
				t.Errorf("MapToHTTPStatus(%q) = %d, want %d", tt.code, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// RespondWithError: AppError integration
// ---------------------------------------------------------------------------

func TestRespondWithError_WithAppError(t *testing.T) {
	c, w := errorTestGinContext("GET", "/api/v1/resource/123")

	appErr := apperrors.NewServiceError(apperrors.ErrServiceNotFound, "resource not found")
	RespondWithError(c, appErr, 0)

	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Status != consts.ResponseStatusError {
		t.Errorf("expected status 'error', got %q", resp.Status)
	}
	if resp.Error.Path != "/api/v1/resource/123" {
		t.Errorf("expected path '/api/v1/resource/123', got %q", resp.Error.Path)
	}
}

func TestRespondWithError_PlainError_FallbackCode(t *testing.T) {
	c, w := errorTestGinContext("POST", "/api/v1/items")

	plainErr := errors.New("something broke")
	RespondWithError(c, plainErr, http.StatusBadGateway)

	if w.Code != http.StatusBadGateway {
		t.Errorf("expected %d, got %d", http.StatusBadGateway, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Error.Message != "An unexpected error occurred." {
		t.Errorf("expected fallback message, got %q", resp.Error.Message)
	}
}

func TestRespondWithError_PlainError_DefaultsTo500(t *testing.T) {
	c, w := errorTestGinContext("DELETE", "/api/v1/items/1")

	plainErr := errors.New("unknown failure")
	RespondWithError(c, plainErr, 0)

	if w.Code != 500 {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Suggestion fallback
// ---------------------------------------------------------------------------

func TestRespondWithError_SuggestionForKnownStatus(t *testing.T) {
	c, w := errorTestGinContext("GET", "/api/v1/test")
	RespondWithError(c, errors.New("not found"), http.StatusNotFound)

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Error.Suggestion == "" {
		t.Error("expected non-empty suggestion for 404")
	}
}

func TestRespondWithError_SuggestionFor5xx(t *testing.T) {
	c, w := errorTestGinContext("GET", "/api/v1/test")
	RespondWithError(c, errors.New("server error"), 501)

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Error.Suggestion == "" {
		t.Error("expected non-empty suggestion for 501")
	}
}

// ---------------------------------------------------------------------------
// Documentation URL
// ---------------------------------------------------------------------------

func TestRespondWithError_DocumentationURL(t *testing.T) {
	c, w := errorTestGinContext("GET", "/api/v1/test")
	RespondWithError(c, errors.New("err"), http.StatusTeapot)

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	expected := DefaultDocsURL + "/418"
	if resp.DocumentationUrl != expected {
		t.Errorf("expected docs URL %q, got %q", expected, resp.DocumentationUrl)
	}
}

// ---------------------------------------------------------------------------
// Accept-Language header
// ---------------------------------------------------------------------------

func TestRespondWithError_AcceptLanguage(t *testing.T) {
	c, w := errorTestGinContext("GET", "/api/v1/test")
	c.Request.Header.Set("Accept-Language", "uz-UZ")

	appErr := apperrors.NewServiceError(apperrors.ErrServiceNotFound, "not found")
	RespondWithError(c, appErr, 0)

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	// We just verify it doesn't crash with non-English; actual message depends on translations
	if resp.Error.Message == "" {
		t.Error("expected non-empty message")
	}
	_ = w // suppress unused
}

// ---------------------------------------------------------------------------
// Timestamp
// ---------------------------------------------------------------------------

func TestRespondWithError_HasTimestamp(t *testing.T) {
	c, w := errorTestGinContext("GET", "/api/v1/test")
	RespondWithError(c, errors.New("err"), 400)

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Error.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

// ---------------------------------------------------------------------------
// statusSuggestions coverage
// ---------------------------------------------------------------------------

func TestStatusSuggestions_AllKeysHaveValues(t *testing.T) {
	for code, suggestion := range statusSuggestions {
		if suggestion == "" {
			t.Errorf("empty suggestion for status code %d", code)
		}
	}
}
