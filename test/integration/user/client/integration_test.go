package client

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

func TestUserAPI_Integration_Exhaustive(t *testing.T) {
	cleanDB(t)
	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg)

	handler := gin.New()
	restapi.NewRouter(handler, setup.TestCfg, useCases, l)
	ctx := t.Context()

	// Pre-seed a user for some tests
	seededPhone := "998901112233"
	const seededPass = "password123"
	u := domain.NewUser()
	u.ID = uuid.New()
	u.Username = stringPtr("seeded_user")
	u.Phone = &seededPhone
	u.SetPassword(seededPass)
	if err := repositories.Persistent.Postgres.User.Client.Create(ctx, u); err != nil {
		t.Fatalf("Failed to seed user: %v", err)
	}
	dbUser, _ := repositories.Persistent.Postgres.User.Client.GetByPhone(ctx, seededPhone)

	// Get token for authenticated tests
	signinBody, _ := json.Marshal(map[string]string{
		"phone":    seededPhone,
		"password": seededPass,
	})
	wLogin := httptest.NewRecorder()
	handler.ServeHTTP(wLogin, httptest.NewRequest(http.MethodPost, "/api/v1/users/sign-in", bytes.NewBuffer(signinBody)))
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

	unicodePhone := "99892" + uuid.New().String()[:7]
	unicodeUsername := "Пользователь_Юникод_" + uuid.New().String()[:6]
	longPhone := "99893" + uuid.New().String()[:7]
	longUsername := "this_is_a_very_long_username_that_should_be_stored_correctly_in_database_" + uuid.New().String()[:6]

	type testCase struct {
		name          string
		method        string
		url           string
		body          any
		useToken      bool
		expectedCode  int
		checkResponse func(t *testing.T, body []byte)
	}

	testCases := []testCase{
		// SIGN UP SCENARIOS (1-6)
		{
			name:   "SIGNUP SUCCESS: Normal registration",
			method: http.MethodPost,
			url:    "/api/v1/users/sign-up",
			body: map[string]string{
				"username": "user1",
				"phone":    "998900000001",
				"password": "password123",
			},
			expectedCode: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var resp map[string]any
				json.Unmarshal(body, &resp)
				data, _ := resp["data"].(map[string]any)
				assert.NotEmpty(t, data["access_token"])

				// DB Check
				dbU, err := repositories.Persistent.Postgres.User.Client.GetByPhone(ctx, "998900000001")
				assert.NoError(t, err)
				if assert.NotNil(t, dbU) && assert.NotNil(t, dbU.Username) {
					assert.Equal(t, "user1", *dbU.Username)
				}
			},
		},
		{
			name:   "SIGNUP CONFLICT: Duplicate phone",
			method: http.MethodPost,
			url:    "/api/v1/users/sign-up",
			body: map[string]string{
				"username": "user2",
				"phone":    seededPhone,
				"password": "password123",
			},
			expectedCode: http.StatusConflict,
		},
		{
			name:   "SIGNUP VALIDATION: Missing phone",
			method: http.MethodPost,
			url:    "/api/v1/users/sign-up",
			body: map[string]string{
				"username": "user_no_phone",
				"password": "password123",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:   "SIGNUP SUCCESS: Unicode username",
			method: http.MethodPost,
			url:    "/api/v1/users/sign-up",
			body: map[string]string{
				"username": unicodeUsername,
				"phone":    unicodePhone,
				"password": "password123",
			},
			expectedCode: http.StatusCreated,

			checkResponse: func(t *testing.T, body []byte) {
				var resp map[string]any
				json.Unmarshal(body, &resp)
				data, _ := resp["data"].(map[string]any)
				assert.NotEmpty(t, data["access_token"])

				// DB Check
				dbU, err := repositories.Persistent.Postgres.User.Client.GetByPhone(ctx, unicodePhone)
				assert.NoError(t, err)
				if assert.NotNil(t, dbU) && assert.NotNil(t, dbU.Username) {
					assert.Equal(t, unicodeUsername, *dbU.Username)
				}
			},
		},
		{
			name:   "SIGNUP SUCCESS: Long username",
			method: http.MethodPost,
			url:    "/api/v1/users/sign-up",
			body: map[string]string{
				"username": longUsername,
				"phone":    longPhone,
				"password": "password",
			},
			expectedCode: http.StatusCreated,

			checkResponse: func(t *testing.T, body []byte) {
				var resp map[string]any
				json.Unmarshal(body, &resp)
				data, _ := resp["data"].(map[string]any)
				assert.NotEmpty(t, data["access_token"])

				// DB Check
				dbU, err := repositories.Persistent.Postgres.User.Client.GetByPhone(ctx, longPhone)
				assert.NoError(t, err)
				if assert.NotNil(t, dbU) && assert.NotNil(t, dbU.Username) {
					assert.Equal(t, longUsername, *dbU.Username)
				}
			},
		},
		{
			name:         "SIGNUP FAIL: Empty body",
			method:       http.MethodPost,
			url:          "/api/v1/users/sign-up",
			body:         nil,
			expectedCode: http.StatusBadRequest,
		},

		// SIGN IN SCENARIOS (7-12)
		{
			name:   "SIGNIN SUCCESS: Valid credentials",
			method: http.MethodPost,
			url:    "/api/v1/users/sign-in",
			body: map[string]string{
				"phone":    seededPhone,
				"password": seededPass,
			},
			expectedCode: http.StatusOK,

			checkResponse: func(t *testing.T, body []byte) {
				var resp map[string]any
				json.Unmarshal(body, &resp)
				assert.NotEmpty(t, resp["data"].(map[string]any)["access_token"])
			},
		},
		{
			name:   "SIGNIN FAIL: Wrong password",
			method: http.MethodPost,
			url:    "/api/v1/users/sign-in",
			body: map[string]string{
				"phone":    seededPhone,
				"password": "wrong_password",
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:   "SIGNIN FAIL: Non-existent user",
			method: http.MethodPost,
			url:    "/api/v1/users/sign-in",
			body: map[string]string{
				"phone":    "998909876543",
				"password": "any",
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:   "SIGNIN FAIL: Empty phone",
			method: http.MethodPost,
			url:    "/api/v1/users/sign-in",
			body: map[string]string{
				"password": seededPass,
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "SIGNIN FAIL: Invalid JSON",
			method:       http.MethodPost,
			url:          "/api/v1/users/sign-in",
			body:         "{not_json}",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:   "SIGNIN SUCCESS: Re-signin multiple times",
			method: http.MethodPost,
			url:    "/api/v1/users/sign-in",
			body: map[string]string{
				"phone":    seededPhone,
				"password": seededPass,
			},
			expectedCode: http.StatusOK,
		},

		// AUTHENTICATED USER OPERATIONS (13-20)
		{
			name:         "USER GET: Success fetching profile",
			method:       http.MethodGet,
			url:          "/api/v1/users/" + dbUser.ID.String(),
			useToken:     true,
			expectedCode: http.StatusOK,

			checkResponse: func(t *testing.T, body []byte) {
				var resp map[string]any
				json.Unmarshal(body, &resp)
				assert.NotNil(t, resp["data"], "Response data should not be nil")
				data, ok := resp["data"].(map[string]any)
				assert.True(t, ok, "Data should be a map")
				assert.Equal(t, "seeded_user", data["username"])
			},
		},
		{
			name:         "USER GET: Unauthorized (no token)",
			method:       http.MethodGet,
			url:          "/api/v1/users/" + dbUser.ID.String(),
			useToken:     false,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "USER PATCH: Update username success",
			method:       http.MethodPatch,
			url:          "/api/v1/users/" + dbUser.ID.String(),
			body:         map[string]string{"username": "new_seeded_name"},
			useToken:     true,
			expectedCode: http.StatusOK,

			checkResponse: func(t *testing.T, body []byte) {
				var resp map[string]any
				json.Unmarshal(body, &resp)
				assert.Equal(t, "SUCCESS", resp["status"])

				// DB Check
				dbU, err := repositories.Persistent.Postgres.User.Client.Get(ctx, &domain.UserFilter{ID: &dbUser.ID})
				assert.NoError(t, err)
				if assert.NotNil(t, dbU) && assert.NotNil(t, dbU.Username) {
					assert.Equal(t, "new_seeded_name", *dbU.Username)
				}
			},
		},
		{
			name:         "USER PATCH: Unauthorized update",
			method:       http.MethodPatch,
			url:          "/api/v1/users/" + dbUser.ID.String(),
			body:         map[string]string{"username": "hacker"},
			useToken:     false,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "USER LIST: Success listing multiple users",
			method:       http.MethodGet,
			url:          "/api/v1/users/",
			useToken:     true,
			expectedCode: http.StatusOK,

			checkResponse: func(t *testing.T, body []byte) {
				var resp map[string]any
				json.Unmarshal(body, &resp)
				assert.NotNil(t, resp["data"], "Response data should not be nil")
				data, ok := resp["data"].([]any)
				assert.True(t, ok, "Data should be a slice")
				assert.GreaterOrEqual(t, len(data), 1)
			},
		},
		{
			name:         "USER DELETE: Success",
			method:       http.MethodDelete,
			url:          "/api/v1/users/" + dbUser.ID.String(),
			useToken:     true,
			expectedCode: http.StatusOK,

			checkResponse: func(t *testing.T, body []byte) {
				var resp map[string]any
				json.Unmarshal(body, &resp)
				assert.Equal(t, "SUCCESS", resp["status"])

				// DB Check
				_, err := repositories.Persistent.Postgres.User.Client.Get(ctx, &domain.UserFilter{ID: &dbUser.ID})
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "record not found")
			},
		},
		{
			name:         "USER GET: Post-delete not found",
			method:       http.MethodGet,
			url:          "/api/v1/users/" + dbUser.ID.String(),
			useToken:     true,                // Token still valid for check but user gone
			expectedCode: http.StatusNotFound, // Correctly returns 404 for missing record

			checkResponse: func(t *testing.T, body []byte) {
				var resp map[string]any
				json.Unmarshal(body, &resp)
				assert.NotNil(t, resp["error"])
			},
		},
		{
			name:         "USER SIGNOUT: Success",
			method:       http.MethodPost,
			url:          "/api/v1/users/sign-out",
			useToken:     true,
			expectedCode: http.StatusOK,

			checkResponse: func(t *testing.T, body []byte) {
				var resp map[string]any
				json.Unmarshal(body, &resp)
				assert.Equal(t, "SUCCESS", resp["status"])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var bodyReader *bytes.Buffer
			if tc.body != nil {
				if s, ok := tc.body.(string); ok {
					bodyReader = bytes.NewBufferString(s)
				} else {
					jsonBody, _ := json.Marshal(tc.body)
					bodyReader = bytes.NewBuffer(jsonBody)
				}
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
