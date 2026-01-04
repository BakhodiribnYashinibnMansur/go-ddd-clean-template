package minio

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"gct/internal/controller/restapi"
	"gct/internal/repo"
	"gct/internal/usecase"
	"gct/pkg/logger"
	"gct/test/integration/common/setup"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMinioAPI_Integration_TableDriven(t *testing.T) {
	cleanDB(t)
	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, nil, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg)

	handler := gin.New()
	restapi.NewRouter(handler, setup.TestCfg, useCases, l)

	// Setup user and token
	signupBody, _ := json.Marshal(map[string]string{
		"username": "minio_tester",
		"phone":    "998909990001",
		"password": "password",
	})
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/api/user/users/sign-up", bytes.NewBuffer(signupBody)))

	signinBody, _ := json.Marshal(map[string]string{"phone": "998909990001", "password": "password"})
	wL := httptest.NewRecorder()
	handler.ServeHTTP(wL, httptest.NewRequest(http.MethodPost, "/api/user/users/sign-in", bytes.NewBuffer(signinBody)))
	var loginResp map[string]any
	json.Unmarshal(wL.Body.Bytes(), &loginResp)
	token := loginResp["data"].(map[string]any)["access_token"].(string)

	type testCase struct {
		name          string
		method        string
		url           string
		setupBody     func() ([]byte, string)
		useToken      bool
		expectedCode  int
		checkResponse func(t *testing.T, w *httptest.ResponseRecorder)
		definition    string
	}

	testCases := []testCase{
		{
			name:   "SUCCESS: Upload Image",
			method: http.MethodPost,
			url:    "/api/v1/files/upload/image",
			setupBody: func() ([]byte, string) {
				b := &bytes.Buffer{}
				w := multipart.NewWriter(b)
				p, _ := w.CreateFormFile("file", "test.jpg")
				p.Write([]byte("fake-image"))
				w.Close()
				return b.Bytes(), w.FormDataContentType()
			},
			useToken:     true,
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				t.Helper()
				var resp map[string]any
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NotEmpty(t, resp["data"])
			},
			definition: "Integration test: verifies image upload endpoint works correctly",
		},
		{
			name:   "SUCCESS: Upload Multiple Images",
			method: http.MethodPost,
			url:    "/api/v1/files/upload/images",
			setupBody: func() ([]byte, string) {
				b := &bytes.Buffer{}
				w := multipart.NewWriter(b)
				p1, _ := w.CreateFormFile("files", "1.jpg")
				p1.Write([]byte("img1"))
				p2, _ := w.CreateFormFile("files", "2.png")
				p2.Write([]byte("img2"))
				w.Close()
				return b.Bytes(), w.FormDataContentType()
			},
			useToken:     true,
			expectedCode: http.StatusOK,
			definition:   "Integration test: verifies batch image upload endpoint handles multiple files",
		},
		{
			name:   "SUCCESS: Upload Doc",
			method: http.MethodPost,
			url:    "/api/v1/files/upload/doc",
			setupBody: func() ([]byte, string) {
				b := &bytes.Buffer{}
				w := multipart.NewWriter(b)
				p, _ := w.CreateFormFile("doc", "test.pdf")
				p.Write([]byte("pdf-content"))
				w.Close()
				return b.Bytes(), w.FormDataContentType()
			},
			useToken:     true,
			expectedCode: http.StatusOK,
			definition:   "Integration test: verifies document upload endpoint works correctly",
		},
		{
			name:   "SUCCESS: Upload Video",
			method: http.MethodPost,
			url:    "/api/v1/files/upload/video",
			setupBody: func() ([]byte, string) {
				b := &bytes.Buffer{}
				w := multipart.NewWriter(b)
				p, _ := w.CreateFormFile("video", "test.mp4")
				p.Write([]byte("video-content"))
				w.Close()
				return b.Bytes(), w.FormDataContentType()
			},
			useToken:     true,
			expectedCode: http.StatusOK,
			definition:   "Integration test: verifies video upload endpoint works correctly",
		},
		{
			name:   "VALIDATION: Invalid extension",
			method: http.MethodPost,
			url:    "/api/v1/files/upload/image",
			setupBody: func() ([]byte, string) {
				b := &bytes.Buffer{}
				w := multipart.NewWriter(b)
				p, _ := w.CreateFormFile("file", "test.exe")
				p.Write([]byte("binary"))
				w.Close()
				return b.Bytes(), w.FormDataContentType()
			},
			useToken:     true,
			expectedCode: http.StatusBadRequest,
			definition:   "Integration test: ensures invalid file extensions are rejected",
		},
		{
			name:   "VALIDATION: No file",
			method: http.MethodPost,
			url:    "/api/v1/files/upload/image",
			setupBody: func() ([]byte, string) {
				return nil, ""
			},
			useToken:     true,
			expectedCode: http.StatusBadRequest,
			definition:   "Integration test: ensures empty request is rejected",
		},
		{
			name:   "UNAUTHORIZED: No token",
			method: http.MethodPost,
			url:    "/api/v1/files/upload/image",
			setupBody: func() ([]byte, string) {
				b := &bytes.Buffer{}
				w := multipart.NewWriter(b)
				p, _ := w.CreateFormFile("file", "test.jpg")
				p.Write([]byte("data"))
				w.Close()
				return b.Bytes(), w.FormDataContentType()
			},
			useToken:     false,
			expectedCode: http.StatusUnauthorized,
			definition:   "Integration test: verifies authentication is required for uploads",
		},
		{
			name:   "SUCCESS: Download file",
			method: http.MethodGet,
			url:    "/api/v1/files/download",
			setupBody: func() ([]byte, string) {
				tmp, _ := os.CreateTemp(t.TempDir(), "integration-*.txt")
				tmp.WriteString("download-me")
				tmp.Close()
				return nil, "file-path=" + tmp.Name()
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				t.Helper()
				assert.Equal(t, "download-me", w.Body.String())
			},
			definition: "Integration test: verifies file download endpoint works correctly",
		},
		{
			name:         "FAIL: Download non-existent",
			method:       http.MethodGet,
			url:          "/api/v1/files/download?file-path=/tmp/missing",
			expectedCode: http.StatusBadRequest,
			definition:   "Integration test: ensures non-existent files return appropriate error",
		},
		{
			name:         "NOT IMPLEMENTED: Transfer",
			method:       http.MethodPost,
			url:          "/api/v1/files/transfer",
			useToken:     true,
			expectedCode: http.StatusNotImplemented,
			definition:   "Integration test: verifies transfer endpoint returns NotImplemented status",
		},
		{
			name:         "METHOD NOT ALLOWED: Post to download",
			method:       http.MethodPost,
			url:          "/api/v1/files/download",
			expectedCode: http.StatusMethodNotAllowed,
			definition:   "Integration test: ensures download endpoint only accepts GET requests",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var body []byte
			var contentType string
			url := tc.url

			if tc.setupBody != nil {
				b, ct := tc.setupBody()
				body = b
				if ct != "" && (tc.method == http.MethodPost || tc.method == http.MethodPatch) {
					contentType = ct
				} else if ct != "" && tc.method == http.MethodGet {
					url = url + "?" + ct
				}
			}

			req := httptest.NewRequest(tc.method, url, bytes.NewBuffer(body))
			if contentType != "" {
				req.Header.Set("Content-Type", contentType)
			}
			if tc.useToken {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code, "Test Case: %s", tc.name)
			if tc.checkResponse != nil {
				tc.checkResponse(t, w)
			}
		})
	}
}
