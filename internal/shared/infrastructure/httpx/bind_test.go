package httpx

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type testBindRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

type testQueryRequest struct {
	Page  int    `form:"page" binding:"required"`
	Query string `form:"query"`
}

type testURIRequest struct {
	ID string `uri:"id" binding:"required"`
}

func createTestGinContext(method, path, body string, headers map[string]string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	c.Request = req
	return c, w
}

func TestBindAndValidate_Success(t *testing.T) {
	c, _ := createTestGinContext("POST", "/test", `{"name":"John","email":"john@example.com"}`, map[string]string{
		"Content-Type": "application/json",
	})

	var req testBindRequest
	ok := BindAndValidate(c, &req)
	if !ok {
		t.Fatal("expected BindAndValidate to return true for valid input")
	}
	if req.Name != "John" {
		t.Errorf("expected name 'John', got %q", req.Name)
	}
	if req.Email != "john@example.com" {
		t.Errorf("expected email 'john@example.com', got %q", req.Email)
	}
}

func TestBindAndValidate_InvalidInput(t *testing.T) {
	c, _ := createTestGinContext("POST", "/test", `{"name":""}`, map[string]string{
		"Content-Type": "application/json",
	})

	var req testBindRequest
	ok := BindAndValidate(c, &req)
	if ok {
		t.Fatal("expected BindAndValidate to return false for invalid input")
	}
}

func TestBindAndValidate_MalformedJSON(t *testing.T) {
	c, _ := createTestGinContext("POST", "/test", `{invalid`, map[string]string{
		"Content-Type": "application/json",
	})

	var req testBindRequest
	ok := BindAndValidate(c, &req)
	if ok {
		t.Fatal("expected BindAndValidate to return false for malformed JSON")
	}
}

func TestBindJSON_Success(t *testing.T) {
	c, _ := createTestGinContext("POST", "/test", `{"name":"Jane","email":"jane@test.com"}`, map[string]string{
		"Content-Type": "application/json",
	})

	var req testBindRequest
	ok := BindJSON(c, &req)
	if !ok {
		t.Fatal("expected BindJSON to return true for valid JSON input")
	}
	if req.Name != "Jane" {
		t.Errorf("expected name 'Jane', got %q", req.Name)
	}
}

func TestBindJSON_InvalidInput(t *testing.T) {
	c, _ := createTestGinContext("POST", "/test", `{"name":""}`, map[string]string{
		"Content-Type": "application/json",
	})

	var req testBindRequest
	ok := BindJSON(c, &req)
	if ok {
		t.Fatal("expected BindJSON to return false for invalid input")
	}
}

func TestBindQuery_Success(t *testing.T) {
	c, _ := createTestGinContext("GET", "/test?page=1&query=hello", "", nil)

	var req testQueryRequest
	ok := BindQuery(c, &req)
	if !ok {
		t.Fatal("expected BindQuery to return true for valid query params")
	}
	if req.Page != 1 {
		t.Errorf("expected page 1, got %d", req.Page)
	}
	if req.Query != "hello" {
		t.Errorf("expected query 'hello', got %q", req.Query)
	}
}

func TestBindQuery_MissingRequired(t *testing.T) {
	c, _ := createTestGinContext("GET", "/test?query=hello", "", nil)

	var req testQueryRequest
	ok := BindQuery(c, &req)
	if ok {
		t.Fatal("expected BindQuery to return false when required field missing")
	}
}

func TestBindURI_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test/123", nil)
	c.Params = gin.Params{{Key: "id", Value: "123"}}

	var req testURIRequest
	ok := BindURI(c, &req)
	if !ok {
		t.Fatal("expected BindURI to return true for valid URI params")
	}
	if req.ID != "123" {
		t.Errorf("expected ID '123', got %q", req.ID)
	}
}

func TestBindURI_MissingRequired(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test/", nil)
	c.Params = gin.Params{}

	var req testURIRequest
	ok := BindURI(c, &req)
	if ok {
		t.Fatal("expected BindURI to return false when required URI param missing")
	}
}
