package session

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/controller/restapi"
	"gct/internal/domain"
	"gct/internal/repo"
	"gct/internal/usecase"
	"gct/pkg/logger"
	"gct/test/integration/common/setup"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSessionAPI_Integration(t *testing.T) {
	cleanDB(t)
	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	ctx := t.Context()

	handler := gin.New()
	restapi.NewRouter(handler, setup.TestCfg, useCases, l)

	// Setup user and a session
	token, sessionID := createUserAndSession(t, handler, "998908001001", "password123")

	type testCase struct {
		name          string
		method        string
		url           string
		body          any
		useToken      bool
		expectedCode  int
		checkResponse func(t *testing.T, body []byte)
		definition    string
	}

	testCases := []testCase{
		{
			name:         "SUCCESS: List sessions",
			method:       http.MethodGet,
			url:          "/api/v1/sessions/",
			useToken:     true,
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				json.Unmarshal(body, &resp)
				assert.NotNil(t, resp["data"])
				sessions := resp["data"].([]any)
				assert.GreaterOrEqual(t, len(sessions), 1)
			},
			definition: "Verifies authenticated user can list their sessions",
		},
		{
			name:         "SUCCESS: Get session by ID",
			method:       http.MethodGet,
			url:          "/api/v1/sessions/" + sessionID,
			useToken:     true,
			expectedCode: http.StatusOK,

			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				json.Unmarshal(body, &resp)
				assert.Equal(t, sessionID, resp["data"].(map[string]any)["id"])
			},
			definition: "Verifies authenticated user can get specific session details",
		},
		{
			name:         "SUCCESS: Update session activity",
			method:       http.MethodPatch,
			url:          "/api/v1/sessions/" + sessionID + "/activity",
			useToken:     true,
			expectedCode: http.StatusOK,

			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				json.Unmarshal(body, &resp)
				assert.Equal(t, "SUCCESS", resp["status"])

				// DB Check
				sid, _ := uuid.Parse(sessionID)
				dbS, err := repositories.Persistent.Postgres.User.SessionRepo.Get(ctx, &domain.SessionFilter{ID: &sid})
				assert.NoError(t, err)
				assert.NotNil(t, dbS.LastActivity)
			},
			definition: "Verifies session activity timestamp can be updated",
		},
		{
			name:         "SUCCESS: Delete session",
			method:       http.MethodDelete,
			url:          "/api/v1/sessions/" + sessionID,
			useToken:     true,
			expectedCode: http.StatusOK,

			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				json.Unmarshal(body, &resp)
				assert.Equal(t, "SUCCESS", resp["status"])

				// DB Check
				sid, _ := uuid.Parse(sessionID)
				_, err := repositories.Persistent.Postgres.User.SessionRepo.Get(ctx, &domain.SessionFilter{ID: &sid})
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "record not found")
			},
			definition: "Verifies session can be manually deleted",
		},
		{
			name:         "NOT FOUND: Get deleted session",
			method:       http.MethodGet,
			url:          "/api/v1/sessions/" + sessionID,
			useToken:     true,
			expectedCode: http.StatusNotFound,
			definition:   "Ensures deleted session is no longer accessible",
		},
		{
			name:         "UNAUTHORIZED: List sessions without token",
			method:       http.MethodGet,
			url:          "/api/v1/sessions/",
			useToken:     false,
			expectedCode: http.StatusUnauthorized,
			definition:   "Ensures authentication is required for listing sessions",
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

			req := httptest.NewRequest(tc.method, tc.url, bodyReader)
			req.Header.Set("Content-Type", "application/json")
			if tc.useToken {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code, "Test Case: %s", tc.name)
			if tc.checkResponse != nil {
				tc.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func createUserAndSession(t *testing.T, handler *gin.Engine, phone, password string) (string, string) {
	// 1. Sign up
	signupBody, _ := json.Marshal(map[string]string{
		"username": "user_" + phone,
		"phone":    phone,
		"password": password,
	})
	wSignup := httptest.NewRecorder()
	handler.ServeHTTP(wSignup,
		httptest.NewRequest(http.MethodPost, "/api/v1/users/sign-up", bytes.NewBuffer(signupBody)))
	if wSignup.Code != http.StatusCreated && wSignup.Code != http.StatusConflict {
		t.Fatalf("Sign-up failed with status %d: %s", wSignup.Code, wSignup.Body.String())
	}

	// 2. Sign in
	signinBody, _ := json.Marshal(map[string]string{
		"phone":    phone,
		"password": password,
	})
	wLogin := httptest.NewRecorder()
	handler.ServeHTTP(wLogin,
		httptest.NewRequest(http.MethodPost, "/api/v1/users/sign-in", bytes.NewBuffer(signinBody)))

	if wLogin.Code != http.StatusOK {
		t.Fatalf("Sign-in failed with status %d: %s", wLogin.Code, wLogin.Body.String())
	}

	var loginResp map[string]any
	json.Unmarshal(wLogin.Body.Bytes(), &loginResp)
	data, ok := loginResp["data"].(map[string]any)
	if !ok {
		t.Fatalf("Sign-in response data is not a map: %v", loginResp["data"])
	}
	token := data["access_token"].(string)

	// 3. Get user
	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	ctx := t.Context()
	user, _ := repositories.Persistent.Postgres.User.Client.GetByPhone(ctx, phone)

	// 4. Create session
	sessionBody, _ := json.Marshal(map[string]any{
		"user_id":            user.ID,
		"refresh_token_hash": "refresh_" + phone,
		"user_agent":         "IntegrationTest/1.0",
		"ip_address":         "127.0.0.1",
	})
	wSession := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions/", bytes.NewBuffer(sessionBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	handler.ServeHTTP(wSession, req)

	var sessionResp map[string]any
	json.Unmarshal(wSession.Body.Bytes(), &sessionResp)
	sessionID := sessionResp["data"].(map[string]any)["id"].(string)

	return token, sessionID
}
