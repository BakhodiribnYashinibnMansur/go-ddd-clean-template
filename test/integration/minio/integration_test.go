package minio

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/consts"
	minioController "gct/internal/controller/restapi/v1/minio"
	"gct/internal/domain"
	"gct/internal/repo"
	"gct/internal/usecase"
	"gct/pkg/logger"
	"gct/test/integration/common/setup"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMinioAPI_Integration_Direct(t *testing.T) {
	cleanDB(t)
	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	ctx := t.Context()

	controller := minioController.New(useCases, l)

	// Pre-seed user for auth context
	u := domain.NewUser()
	u.ID = uuid.New()
	u.Username = stringPtr("minio_tester")
	phone := "998909990001"
	u.Phone = &phone
	u.SetPassword("password")
	repositories.Persistent.Postgres.User.Client.Create(ctx, u)

	// Helper for auth context
	createAuthContext := func(w *httptest.ResponseRecorder, r *http.Request) *gin.Context {
		c, _ := gin.CreateTestContext(w)
		c.Request = r

		sess := &domain.Session{
			ID:        uuid.New(),
			UserID:    u.ID,
			IPAddress: stringPtr("127.0.0.1"),
			UserAgent: stringPtr("test-agent"),
			ExpiresAt: time.Now().Add(time.Hour),
			CreatedAt: time.Now(),
		}
		repositories.Persistent.Postgres.User.SessionRepo.Create(ctx, sess)

		c.Set(consts.CtxSessionID, sess.ID)
		c.Set(consts.CtxUserID, u.ID.String())
		c.Set(consts.CtxSession, sess)

		return c
	}

	type testCase struct {
		name          string
		handlerFunc   func(c *gin.Context)
		method        string
		setupReq      func() (*http.Request, *gin.Params) // return req and params if needed
		authenticated bool
		expectedCode  int
		checkResponse func(t *testing.T, w *httptest.ResponseRecorder)
	}

	testCases := []testCase{
		{
			name:        "SUCCESS: Upload Image",
			handlerFunc: controller.UploadImage,
			method:      http.MethodPost,
			setupReq: func() (*http.Request, *gin.Params) {
				b := &bytes.Buffer{}
				w := multipart.NewWriter(b)
				p, _ := w.CreateFormFile("file", "test.jpg")
				p.Write([]byte("fake-image-content"))
				w.Close()
				req := httptest.NewRequest(http.MethodPost, "/", b)
				req.Header.Set("Content-Type", w.FormDataContentType())
				return req, nil
			},
			authenticated: true,
			expectedCode:  http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				t.Helper()
				var resp map[string]any
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NotEmpty(t, resp["data"])
			},
		},
		{
			name:        "SUCCESS: Upload Multiple Images",
			handlerFunc: controller.UploadImages,
			method:      http.MethodPost,
			setupReq: func() (*http.Request, *gin.Params) {
				b := &bytes.Buffer{}
				w := multipart.NewWriter(b)
				p1, _ := w.CreateFormFile("files", "1.jpg")
				p1.Write([]byte("img1"))
				p2, _ := w.CreateFormFile("files", "2.png")
				p2.Write([]byte("img2"))
				w.Close()
				req := httptest.NewRequest(http.MethodPost, "/", b)
				req.Header.Set("Content-Type", w.FormDataContentType())
				return req, nil
			},
			authenticated: true,
			expectedCode:  http.StatusOK,
		},
		{
			name:        "SUCCESS: Upload Doc",
			handlerFunc: controller.UploadDoc,
			method:      http.MethodPost,
			setupReq: func() (*http.Request, *gin.Params) {
				b := &bytes.Buffer{}
				w := multipart.NewWriter(b)
				p, _ := w.CreateFormFile("file", "test.pdf") // also correcting form field 'doc' -> 'file' as per controller
				p.Write([]byte("pdf-content"))
				w.Close()
				req := httptest.NewRequest(http.MethodPost, "/", b)
				req.Header.Set("Content-Type", w.FormDataContentType())
				return req, nil
			},
			authenticated: true,
			expectedCode:  http.StatusOK,
		},
		{
			name:        "SUCCESS: Upload Video",
			handlerFunc: controller.UploadVideo,
			method:      http.MethodPost,
			setupReq: func() (*http.Request, *gin.Params) {
				b := &bytes.Buffer{}
				w := multipart.NewWriter(b)
				p, _ := w.CreateFormFile("video", "test.mp4")
				p.Write([]byte("video-content"))
				w.Close()
				req := httptest.NewRequest(http.MethodPost, "/", b)
				req.Header.Set("Content-Type", w.FormDataContentType())
				return req, nil
			},
			authenticated: true,
			expectedCode:  http.StatusOK,
		},
		{
			name:        "VALIDATION: Invalid extension",
			handlerFunc: controller.UploadImage,
			method:      http.MethodPost,
			setupReq: func() (*http.Request, *gin.Params) {
				b := &bytes.Buffer{}
				w := multipart.NewWriter(b)
				p, _ := w.CreateFormFile("file", "test.exe")
				p.Write([]byte("binary"))
				w.Close()
				req := httptest.NewRequest(http.MethodPost, "/", b)
				req.Header.Set("Content-Type", w.FormDataContentType())
				return req, nil
			},
			authenticated: true,
			expectedCode:  http.StatusBadRequest,
		},
		{
			name:        "VALIDATION: No file",
			handlerFunc: controller.UploadImage,
			method:      http.MethodPost,
			setupReq: func() (*http.Request, *gin.Params) {
				req := httptest.NewRequest(http.MethodPost, "/", nil)
				return req, nil
			},
			authenticated: true,
			expectedCode:  http.StatusBadRequest,
		},
		{
			name:        "NOT IMPLEMENTED: Transfer",
			handlerFunc: controller.TransferFile,
			method:      http.MethodPost,
			setupReq: func() (*http.Request, *gin.Params) {
				req := httptest.NewRequest(http.MethodPost, "/", nil)
				return req, nil
			},
			authenticated: true,
			expectedCode:  http.StatusNotImplemented,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, params := tc.setupReq()

			var c *gin.Context
			if tc.authenticated {
				c = createAuthContext(w, req)
			} else {
				c, _ = gin.CreateTestContext(w)
				c.Request = req
			}

			if params != nil {
				c.Params = *params
			}

			tc.handlerFunc(c)

			assert.Equal(t, tc.expectedCode, w.Code, "Case: %s body: %s", tc.name, w.Body.String())
			if tc.checkResponse != nil {
				tc.checkResponse(t, w)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
