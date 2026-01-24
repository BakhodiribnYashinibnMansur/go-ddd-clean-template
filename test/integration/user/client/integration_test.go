package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/consts"
	clientController "gct/internal/controller/restapi/v1/user/client"
	"gct/internal/domain"
	"gct/internal/repo"
	"gct/internal/usecase"
	"gct/pkg/logger"
	"gct/test/integration/common/setup"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserAPI_Integration_Direct(t *testing.T) {
	cleanDB(t)
	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)

	// Instantiate Controller directly
	controller := clientController.New(useCases, setup.TestCfg, l)
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

	// Helper to create a valid session and token for authenticated requests
	// Since we are mocking middleware/context setting, we need to replicate what middleware does:
	createAuthContext := func(w *httptest.ResponseRecorder, r *http.Request) *gin.Context {
		c, _ := gin.CreateTestContext(w)
		c.Request = r

		// Create a session for seeded user
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

	unicodePhone := "99892" + uuid.New().String()[:7]
	unicodeUsername := "Пользователь_Юникод_" + uuid.New().String()[:6]

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
		// SIGN UP SCENARIOS
		{
			name:        "SIGNUP SUCCESS: Normal registration",
			handlerFunc: controller.SignUp,
			method:      http.MethodPost,
			body: map[string]string{
				"username": "user1",
				"phone":    "998900000001",
				"password": "password123",
			},
			expectedCode: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				json.Unmarshal(body, &resp)
				data, ok := resp["data"].(map[string]any)
				if assert.True(t, ok, "Data should be map") {
					assert.NotEmpty(t, data["access_token"])
				}
				// DB verify
				dbU, err := repositories.Persistent.Postgres.User.Client.GetByPhone(ctx, "998900000001")
				assert.NoError(t, err)
				if assert.NotNil(t, dbU) {
					assert.Equal(t, "user1", *dbU.Username)
				}
			},
		},
		{
			name:        "SIGNUP CONFLICT: Duplicate phone",
			handlerFunc: controller.SignUp,
			method:      http.MethodPost,
			body: map[string]string{
				"username": "user2",
				"phone":    seededPhone,
				"password": "password123",
			},
			expectedCode: http.StatusConflict,
		},
		{
			name:        "SIGNUP SUCCESS: Unicode username",
			handlerFunc: controller.SignUp,
			method:      http.MethodPost,
			body: map[string]string{
				"username": unicodeUsername,
				"phone":    unicodePhone,
				"password": "password123",
			},
			expectedCode: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				dbU, err := repositories.Persistent.Postgres.User.Client.GetByPhone(ctx, unicodePhone)
				assert.NoError(t, err)
				if assert.NotNil(t, dbU) {
					assert.Equal(t, unicodeUsername, *dbU.Username)
				}
			},
		},

		// SIGN IN SCENARIOS
		{
			name:        "SIGNIN SUCCESS: Valid credentials",
			handlerFunc: controller.SignIn,
			method:      http.MethodPost,
			body: map[string]string{
				"phone":    seededPhone,
				"password": seededPass,
			},
			expectedCode: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				json.Unmarshal(body, &resp)
				data, ok := resp["data"].(map[string]any)
				if assert.True(t, ok) {
					assert.NotEmpty(t, data["access_token"])
				}
			},
		},
		{
			name:        "SIGNIN FAIL: Wrong password",
			handlerFunc: controller.SignIn,
			method:      http.MethodPost,
			body: map[string]string{
				"phone":    seededPhone,
				"password": "wrong_password",
			},
			expectedCode: http.StatusUnauthorized,
		},

		// AUTHENTICATED OPERATIONS
		{
			name:          "USER GET: Success fetching profile",
			handlerFunc:   controller.User,
			method:        http.MethodGet,
			params:        gin.Params{{Key: consts.ParamUserID, Value: dbUser.ID.String()}},
			authenticated: true,
			expectedCode:  http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				json.Unmarshal(body, &resp)
				data, ok := resp["data"].(map[string]any)
				assert.True(t, ok)
				assert.Equal(t, "seeded_user", data["username"])
			},
		},
		{
			name:          "USER PATCH: Update username success",
			handlerFunc:   controller.Update,
			method:        http.MethodPatch,
			params:        gin.Params{{Key: consts.ParamUserID, Value: dbUser.ID.String()}},
			body:          map[string]string{"username": "new_seeded_name"},
			authenticated: true,
			expectedCode:  http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				// DB Check
				dbU, err := repositories.Persistent.Postgres.User.Client.Get(ctx, &domain.UserFilter{ID: &dbUser.ID})
				assert.NoError(t, err)
				if assert.NotNil(t, dbU) {
					assert.Equal(t, "new_seeded_name", *dbU.Username)
				}
			},
		},
		// LIST USERS
		{
			name:          "USER LIST: Success",
			handlerFunc:   controller.Users,
			method:        http.MethodGet,
			authenticated: true,
			expectedCode:  http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				json.Unmarshal(body, &resp)
				data, ok := resp["data"].([]any)
				assert.True(t, ok)
				assert.GreaterOrEqual(t, len(data), 1)
			},
		},
		// DELETE USERS
		{
			name:          "USER DELETE: Success",
			handlerFunc:   controller.Delete,
			method:        http.MethodDelete,
			params:        gin.Params{{Key: consts.ParamUserID, Value: dbUser.ID.String()}},
			authenticated: true,
			expectedCode:  http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				t.Helper()
				// DB Verify deleted
				_, err := repositories.Persistent.Postgres.User.Client.Get(ctx, &domain.UserFilter{ID: &dbUser.ID})
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "record not found")
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

			// Invoke
			tc.handlerFunc(c)

			// Assert
			assert.Equal(t, tc.expectedCode, w.Code, "Case: %s body: %s", tc.name, w.Body.String())
			if tc.checkResponse != nil {
				tc.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}
