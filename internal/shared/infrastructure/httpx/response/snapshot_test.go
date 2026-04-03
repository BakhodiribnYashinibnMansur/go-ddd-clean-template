package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/shared/domain/consts"

	"github.com/gin-gonic/gin"
)

// snapshotGinContext creates a minimal gin context for snapshot tests.
func snapshotGinContext(method, path string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, nil)
	return c, w
}

// ---------------------------------------------------------------------------
// TestSuccessResponse_Structure
// ---------------------------------------------------------------------------

func TestSuccessResponse_Structure(t *testing.T) {
	c, w := snapshotGinContext("GET", "/api/v1/items")

	payload := map[string]string{"id": "1", "name": "item"}
	ControllerResponse(c, http.StatusOK, payload, nil, true)

	raw := w.Body.Bytes()
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	requiredKeys := []string{"status", "statusCode", "message", "data"}
	for _, key := range requiredKeys {
		if _, ok := obj[key]; !ok {
			t.Errorf("missing required key %q in success response", key)
		}
	}

	// Verify actual values
	var resp SuccessResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("unmarshal SuccessResponse: %v", err)
	}
	if resp.Status != consts.ResponseStatusSuccess {
		t.Errorf("status = %q, want %q", resp.Status, consts.ResponseStatusSuccess)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("statusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if resp.Message != consts.ResponseMessageSuccess {
		t.Errorf("message = %q, want %q", resp.Message, consts.ResponseMessageSuccess)
	}
	if resp.Data == nil {
		t.Error("data should not be nil")
	}
}

func TestSuccessResponse_OmitsMetaWhenNil(t *testing.T) {
	c, w := snapshotGinContext("GET", "/api/v1/items")
	ControllerResponse(c, http.StatusOK, "ok", nil, true)

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &obj); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	// meta should be omitted when nil
	if raw, ok := obj["meta"]; ok && string(raw) != "null" {
		t.Errorf("meta should be omitted or null when nil, got %s", string(raw))
	}
}

// ---------------------------------------------------------------------------
// TestErrorResponse_Structure
// ---------------------------------------------------------------------------

func TestErrorResponse_Structure(t *testing.T) {
	c, w := snapshotGinContext("POST", "/api/v1/users")
	RespondWithError(c, &simpleError{msg: "not found"}, http.StatusNotFound)

	raw := w.Body.Bytes()
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	topLevelKeys := []string{"status", "statusCode", "error", "requestId", "documentation_url"}
	for _, key := range topLevelKeys {
		if _, ok := obj[key]; !ok {
			t.Errorf("missing required top-level key %q in error response", key)
		}
	}

	// Verify nested error object keys
	var errObj map[string]json.RawMessage
	if err := json.Unmarshal(obj["error"], &errObj); err != nil {
		t.Fatalf("error field is not a valid JSON object: %v", err)
	}

	errorKeys := []string{"code", "numeric_code", "message", "timestamp", "path", "method", "retryable"}
	for _, key := range errorKeys {
		if _, ok := errObj[key]; !ok {
			t.Errorf("missing required key %q in error detail", key)
		}
	}

	// Verify values
	var resp ErrorResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("unmarshal ErrorResponse: %v", err)
	}
	if resp.Status != consts.ResponseStatusError {
		t.Errorf("status = %q, want %q", resp.Status, consts.ResponseStatusError)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("statusCode = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
	if resp.Error.Path != "/api/v1/users" {
		t.Errorf("error.path = %q, want %q", resp.Error.Path, "/api/v1/users")
	}
	if resp.Error.Method != "POST" {
		t.Errorf("error.method = %q, want %q", resp.Error.Method, "POST")
	}
	if resp.RequestId == "" {
		t.Error("requestId should not be empty")
	}
	if resp.DocumentationUrl == "" {
		t.Error("documentation_url should not be empty")
	}
}

// ---------------------------------------------------------------------------
// TestPaginatedResponse_Structure
// ---------------------------------------------------------------------------

func TestPaginatedResponse_Structure(t *testing.T) {
	c, w := snapshotGinContext("GET", "/api/v1/items")

	items := []map[string]string{
		{"id": "1", "name": "a"},
		{"id": "2", "name": "b"},
	}
	meta := Meta{Total: 50, Limit: 10, Offset: 0, Page: 1}
	ControllerResponse(c, http.StatusOK, items, meta, true)

	raw := w.Body.Bytes()
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// data must be present and be an array
	if _, ok := obj["data"]; !ok {
		t.Fatal("missing 'data' key in paginated response")
	}
	var dataSlice []json.RawMessage
	if err := json.Unmarshal(obj["data"], &dataSlice); err != nil {
		t.Fatalf("data is not an array: %v", err)
	}
	if len(dataSlice) != 2 {
		t.Errorf("data length = %d, want 2", len(dataSlice))
	}

	// meta must be present and contain pagination fields
	if _, ok := obj["meta"]; !ok {
		t.Fatal("missing 'meta' key in paginated response")
	}
	var metaObj map[string]json.RawMessage
	if err := json.Unmarshal(obj["meta"], &metaObj); err != nil {
		t.Fatalf("meta is not a valid JSON object: %v", err)
	}

	metaKeys := []string{"total", "limit", "offset", "page"}
	for _, key := range metaKeys {
		if _, ok := metaObj[key]; !ok {
			t.Errorf("missing required meta key %q", key)
		}
	}

	// Verify meta values
	var resp SuccessResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// Meta is deserialized as map[string]any by default
	metaMap, ok := resp.Meta.(map[string]any)
	if !ok {
		t.Fatalf("meta is not a map, got %T", resp.Meta)
	}
	if total, ok := metaMap["total"].(float64); !ok || int64(total) != 50 {
		t.Errorf("meta.total = %v, want 50", metaMap["total"])
	}
	if limit, ok := metaMap["limit"].(float64); !ok || int64(limit) != 10 {
		t.Errorf("meta.limit = %v, want 10", metaMap["limit"])
	}
	if offset, ok := metaMap["offset"].(float64); !ok || int64(offset) != 0 {
		t.Errorf("meta.offset = %v, want 0", metaMap["offset"])
	}
	if page, ok := metaMap["page"].(float64); !ok || int64(page) != 1 {
		t.Errorf("meta.page = %v, want 1", metaMap["page"])
	}
}

// ---------------------------------------------------------------------------
// TestResponse_StatusCodes
// ---------------------------------------------------------------------------

func TestResponse_StatusCodes(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		success bool
		payload any
	}{
		{"200 OK", http.StatusOK, true, "ok"},
		{"201 Created", http.StatusCreated, true, map[string]string{"id": "1"}},
		{"204-like via 200", http.StatusOK, true, nil},
		{"400 Bad Request", http.StatusBadRequest, false, "bad request"},
		{"401 Unauthorized", http.StatusUnauthorized, false, "unauthorized"},
		{"403 Forbidden", http.StatusForbidden, false, "forbidden"},
		{"404 Not Found", http.StatusNotFound, false, "not found"},
		{"409 Conflict", http.StatusConflict, false, "conflict"},
		{"422 Unprocessable", http.StatusUnprocessableEntity, false, "unprocessable"},
		{"500 Internal", http.StatusInternalServerError, false, "internal error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := snapshotGinContext("GET", "/test")
			ControllerResponse(c, tt.code, tt.payload, nil, tt.success)

			if w.Code != tt.code {
				t.Errorf("HTTP status = %d, want %d", w.Code, tt.code)
			}

			if tt.success {
				var resp SuccessResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("unmarshal: %v", err)
				}
				if resp.StatusCode != tt.code {
					t.Errorf("body statusCode = %d, want %d", resp.StatusCode, tt.code)
				}
				if resp.Status != consts.ResponseStatusSuccess {
					t.Errorf("status = %q, want %q", resp.Status, consts.ResponseStatusSuccess)
				}
			} else {
				var resp ErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("unmarshal: %v", err)
				}
				if resp.StatusCode != tt.code {
					t.Errorf("body statusCode = %d, want %d", resp.StatusCode, tt.code)
				}
				if resp.Status != consts.ResponseStatusError {
					t.Errorf("status = %q, want %q", resp.Status, consts.ResponseStatusError)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestMapToHTTPStatus helper coverage
// ---------------------------------------------------------------------------

func TestResponse_MapToHTTPStatus_ServiceCodes(t *testing.T) {
	tests := []struct {
		code string
		want int
	}{
		{"HANDLER_BAD_REQUEST", 400},
		{"HANDLER_UNAUTHORIZED", 401},
		{"HANDLER_FORBIDDEN", 403},
		{"HANDLER_NOT_FOUND", 404},
		{"HANDLER_CONFLICT", 409},
		{"HANDLER_TOO_MANY_REQUESTS", 429},
		{"HANDLER_NOT_IMPLEMENTED", 501},
		{"HANDLER_SERVICE_UNAVAILABLE", 503},
		{"TOTALLY_UNKNOWN", 500},
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
