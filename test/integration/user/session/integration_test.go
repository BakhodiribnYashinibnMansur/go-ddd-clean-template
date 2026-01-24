package session

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/consts"
	sessionController "gct/internal/controller/restapi/v1/user/session"
	"gct/internal/domain"
	"gct/internal/repo"
	"gct/internal/usecase"
	"gct/pkg/logger"
	"gct/test/integration/common/setup"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSessionAPI_Integration_Direct(t *testing.T) {
	cleanDB(t)
	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	ctx := t.Context()

	// Instantiate Controller
	controller := sessionController.New(useCases, l)

	// Pre-seed user
	phone := "998908001001"
	pass := "password123"
	u := domain.NewUser()
	u.ID = uuid.New()
	u.Username = stringPtr("session_test_user")
	u.Phone = &phone
	u.SetPassword(pass)
	repositories.Persistent.Postgres.User.Client.Create(ctx, u)

	// Pre-seed a session for testing operations
	sessionID := uuid.New()
	sess := &domain.Session{
		ID:        sessionID,
		UserID:    u.ID,
		IPAddress: stringPtr("127.0.0.1"),
		UserAgent: stringPtr("test-agent"),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		Revoked:   false,
	}
	repositories.Persistent.Postgres.User.SessionRepo.Create(ctx, sess)

	// Helper to create authenticated context
	createAuthContext := func(w *httptest.ResponseRecorder, r *http.Request) *gin.Context {
		c, _ := gin.CreateTestContext(w)
		c.Request = r

		c.Set(consts.CtxSessionID, sess.ID)
		c.Set(consts.CtxUserID, u.ID.String())
		c.Set(consts.CtxSession, sess)

		return c
	}

	type testCase struct {
		name          string
		handlerFunc   func(c *gin.Context)
		method        string
		params        gin.Params
		body          any
		authenticated bool
		expectedCode  int
		checkResponse func(t *testing.T, body []byte)
	}

	testCases := []testCase{
		{
			name:          "SUCCESS: List sessions",
			handlerFunc:   controller.Sessions,
			method:        http.MethodGet,
			authenticated: true,
			expectedCode:  http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				json.Unmarshal(body, &resp)
				assert.NotNil(t, resp["data"])
				sessions := resp["data"].([]any)
				assert.GreaterOrEqual(t, len(sessions), 1)
			},
		},
		{
			name:          "SUCCESS: Get session by ID",
			handlerFunc:   controller.Session,
			method:        http.MethodGet,
			params:        gin.Params{{Key: consts.ParamSessionID, Value: sessionID.String()}},
			authenticated: true,
			expectedCode:  http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				json.Unmarshal(body, &resp)
				data, ok := resp["data"].(map[string]any)
				if assert.True(t, ok) {
					assert.Equal(t, sessionID.String(), data["id"])
				}
			},
		},
		{
			name:          "SUCCESS: Update session activity",
			handlerFunc:   controller.UpdateActivity,
			method:        http.MethodPatch,
			params:        gin.Params{{Key: consts.ParamSessionID, Value: sessionID.String()}},
			authenticated: true,
			expectedCode:  http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				json.Unmarshal(body, &resp)
				assert.Equal(t, "SUCCESS", resp["status"])

				// DB Check
				dbS, err := repositories.Persistent.Postgres.User.SessionRepo.Get(ctx, &domain.SessionFilter{ID: &sessionID})
				assert.NoError(t, err)
				// LastActivity should be close to now
				assert.WithinDuration(t, time.Now(), dbS.LastActivity, 5*time.Second)
			},
		},
		{
			name:          "SUCCESS: Delete session",
			handlerFunc:   controller.Delete,
			method:        http.MethodDelete,
			params:        gin.Params{{Key: consts.ParamSessionID, Value: sessionID.String()}},
			authenticated: true,
			expectedCode:  http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				json.Unmarshal(body, &resp)
				assert.Equal(t, "SUCCESS", resp["status"])

				// DB Check - should be revoked or deleted? Usually delete endpoint revokes or deletes.
				// Checking implementation of Session.Delete... typically it deletes row or sets revoked.
				// Based on previous test expectation: "record not found" means hard delete.
				_, err := repositories.Persistent.Postgres.User.SessionRepo.Get(ctx, &domain.SessionFilter{ID: &sessionID})
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "record not found")
			},
		},
		{
			name:          "NOT FOUND: Get deleted session",
			handlerFunc:   controller.Session,
			method:        http.MethodGet,
			params:        gin.Params{{Key: consts.ParamSessionID, Value: sessionID.String()}},
			authenticated: true,
			expectedCode:  http.StatusNotFound, // or InternalServerError depending on how Get handles not found
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var bodyReader *bytes.Buffer
			if tc.body != nil {
				jsonBody, _ := json.Marshal(tc.body)
				bodyReader = bytes.NewBuffer(jsonBody)
			} else {
				bodyReader = bytes.NewBuffer(nil)
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, "/", bodyReader)
			req.Header.Set("Content-Type", "application/json")

			var c *gin.Context
			if tc.authenticated {
				c = createAuthContext(w, req)
			} else {
				c, _ = gin.CreateTestContext(w)
				c.Request = req
			}

			if tc.params != nil {
				c.Params = tc.params
			}

			tc.handlerFunc(c)

			// Some controllers might not set status code explicitly if they just return data/error helper
			// Verify code.
			// Note: if controller calls response.ControllerResponse, it sets code.
			if w.Code != tc.expectedCode {
				// Sometimes Not Found returns 404, or 400. Direct repo error might be 500 if not handled.
				// Adjust expectation if needed based on controller logic.
				// For now assert equal.
			}
			assert.Equal(t, tc.expectedCode, w.Code, "Case: %s body: %s", tc.name, w.Body.String())
			if tc.checkResponse != nil {
				tc.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
