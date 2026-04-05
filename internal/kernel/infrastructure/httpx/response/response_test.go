package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/kernel/consts"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func testGinContext(method, path string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, nil)
	return c, w
}

func TestControllerResponse_SuccessWithData(t *testing.T) {
	c, w := testGinContext("GET", "/test")

	data := map[string]string{"key": "value"}
	ControllerResponse(c, http.StatusOK, data, nil, true)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Status != consts.ResponseStatusSuccess {
		t.Errorf("expected status %q, got %q", consts.ResponseStatusSuccess, resp.Status)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected statusCode %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if resp.Message != consts.ResponseMessageSuccess {
		t.Errorf("expected message %q, got %q", consts.ResponseMessageSuccess, resp.Message)
	}
	if resp.Data == nil {
		t.Error("expected non-nil data")
	}
}

func TestControllerResponse_SuccessWithMeta(t *testing.T) {
	c, w := testGinContext("GET", "/test")

	data := []string{"a", "b"}
	meta := Meta{Total: 2, Limit: 10, Offset: 0, Page: 1}
	ControllerResponse(c, http.StatusOK, data, meta, true)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp SuccessResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Meta == nil {
		t.Error("expected non-nil meta")
	}
}

func TestControllerResponse_ErrorWithErrorType(t *testing.T) {
	c, w := testGinContext("GET", "/test")

	testErr := errors.New("something went wrong")
	ControllerResponse(c, http.StatusInternalServerError, testErr, nil, false)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Status != consts.ResponseStatusError {
		t.Errorf("expected status %q, got %q", consts.ResponseStatusError, resp.Status)
	}
}

func TestControllerResponse_ErrorWithString(t *testing.T) {
	c, w := testGinContext("GET", "/test")

	ControllerResponse(c, http.StatusBadRequest, "bad input", nil, false)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Status != consts.ResponseStatusError {
		t.Errorf("expected status %q, got %q", consts.ResponseStatusError, resp.Status)
	}
}

func TestControllerResponse_ErrorWithUnknownPayload(t *testing.T) {
	c, w := testGinContext("GET", "/test")

	ControllerResponse(c, http.StatusInternalServerError, 12345, nil, false)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Error.Message == "" {
		t.Error("expected non-empty error message")
	}
}

func TestRespondWithError(t *testing.T) {
	c, w := testGinContext("POST", "/api/v1/users")

	testErr := errors.New("not found")
	RespondWithError(c, testErr, http.StatusNotFound)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Status != consts.ResponseStatusError {
		t.Errorf("expected status %q, got %q", consts.ResponseStatusError, resp.Status)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected statusCode %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
	if resp.Error.Path != "/api/v1/users" {
		t.Errorf("expected path '/api/v1/users', got %q", resp.Error.Path)
	}
	if resp.Error.Method != "POST" {
		t.Errorf("expected method 'POST', got %q", resp.Error.Method)
	}
	if resp.RequestId == "" {
		t.Error("expected non-empty request ID")
	}
	if resp.DocumentationUrl == "" {
		t.Error("expected non-empty documentation URL")
	}
}

func TestRespondWithError_UsesRequestIDHeader(t *testing.T) {
	c, w := testGinContext("GET", "/test")
	c.Request.Header.Set(consts.HeaderXRequestID, "custom-req-id")

	RespondWithError(c, errors.New("err"), http.StatusBadRequest)

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.RequestId != "custom-req-id" {
		t.Errorf("expected request ID 'custom-req-id', got %q", resp.RequestId)
	}
}

func TestMapToHTTPStatus_HandlerCodes(t *testing.T) {
	tests := []struct {
		code     string
		expected int
	}{
		{"HANDLER_BAD_REQUEST", 400},
		{"HANDLER_UNAUTHORIZED", 401},
		{"HANDLER_FORBIDDEN", 403},
		{"HANDLER_NOT_FOUND", 404},
		{"HANDLER_CONFLICT", 409},
		{"HANDLER_TOO_MANY_REQUESTS", 429},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := MapToHTTPStatus(tt.code)
			if got != tt.expected {
				t.Errorf("MapToHTTPStatus(%q) = %d, want %d", tt.code, got, tt.expected)
			}
		})
	}
}

func TestMapToHTTPStatus_UnknownCode(t *testing.T) {
	got := MapToHTTPStatus("COMPLETELY_UNKNOWN_CODE")
	if got != 500 {
		t.Errorf("expected 500 for unknown code, got %d", got)
	}
}

func TestSimpleError(t *testing.T) {
	e := &simpleError{msg: "test error"}
	if e.Error() != "test error" {
		t.Errorf("expected 'test error', got %q", e.Error())
	}
}

func TestMeta(t *testing.T) {
	m := Meta{
		Total:  100,
		Limit:  10,
		Offset: 20,
		Page:   3,
	}
	if m.Total != 100 {
		t.Errorf("expected Total 100, got %d", m.Total)
	}
	if m.Limit != 10 {
		t.Errorf("expected Limit 10, got %d", m.Limit)
	}
	if m.Offset != 20 {
		t.Errorf("expected Offset 20, got %d", m.Offset)
	}
	if m.Page != 3 {
		t.Errorf("expected Page 3, got %d", m.Page)
	}
}
