package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/shared/domain/consts"
	authController "gct/internal/controller/restapi/v1/authz/auth"
	clientController "gct/internal/controller/restapi/v1/user/client"
	"gct/internal/domain"
	"gct/internal/repo"
	"gct/internal/usecase"
	"gct/internal/shared/infrastructure/logger"
	"gct/test/integration/common/setup"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Auth (SignUp / SignIn / SignOut) — Direct Controller Tests
// ---------------------------------------------------------------------------

func TestSignUp_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	authCtl := authController.New(useCases, setup.TestCfg, l)

	tests := []struct {
		name         string
		body         map[string]any
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name: "success",
			body: map[string]any{
				"username": "signup_user",
				"phone":    "998900000001",
				"password": "Str0ng!Pass",
				"session":  map[string]any{},
			},
			expectedCode: http.StatusCreated,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				data := resp["data"].(map[string]any)
				assert.NotEmpty(t, data["access_token"])
				assert.NotEmpty(t, data["user_id"])
				assert.NotEmpty(t, data["session_id"])

				// DB verification
				ctx := t.Context()
				dbUser, err := repositories.Persistent.Postgres.User.Client.GetByPhone(ctx, "998900000001")
				require.NoError(t, err)
				require.NotNil(t, dbUser)
				assert.Equal(t, "signup_user", *dbUser.Username)
			},
		},
		{
			name: "conflict - duplicate phone",
			body: map[string]any{
				"username": "another_user",
				"phone":    "998900000001",
				"password": "Str0ng!Pass",
				"session":  map[string]any{},
			},
			expectedCode: http.StatusConflict,
		},
		{
			name: "bad request - missing session",
			body: map[string]any{
				"username": "no_session_user",
				"phone":    "998900000099",
				"password": "Str0ng!Pass",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "unicode username",
			body: map[string]any{
				"username": "Пользователь_Юникод",
				"phone":    "998900000002",
				"password": "Str0ng!Pass",
				"session":  map[string]any{},
			},
			expectedCode: http.StatusCreated,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				ctx := t.Context()
				dbUser, err := repositories.Persistent.Postgres.User.Client.GetByPhone(ctx, "998900000002")
				require.NoError(t, err)
				assert.Equal(t, "Пользователь_Юникод", *dbUser.Username)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodPost, tt.body)
			authCtl.SignUp(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

func TestSignIn_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	authCtl := authController.New(useCases, setup.TestCfg, l)
	ctx := t.Context()

	// Seed a user via direct signup
	phone := "998901112233"
	password := "Str0ng!Pass"
	u := domain.NewUser()
	u.ID = uuid.New()
	u.Username = stringPtr("signin_user")
	u.Phone = &phone
	u.IsApproved = true
	require.NoError(t, u.SetPassword(password))
	require.NoError(t, repositories.Persistent.Postgres.User.Client.Create(ctx, u))

	tests := []struct {
		name         string
		body         map[string]any
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name: "success",
			body: map[string]any{
				"login":    phone,
				"password": password,
				"session":  map[string]any{},
			},
			expectedCode: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				data := resp["data"].(map[string]any)
				assert.NotEmpty(t, data["access_token"])
				assert.NotEmpty(t, data["refresh_token"])
				assert.Equal(t, u.ID.String(), data["user_id"])
			},
		},
		{
			name: "wrong password",
			body: map[string]any{
				"login":    phone,
				"password": "Wr0ng!Pass",
				"session":  map[string]any{},
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "user not found",
			body: map[string]any{
				"login":    "998909999999",
				"password": "Str0ng!Pass",
				"session":  map[string]any{},
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "bad request - missing session",
			body: map[string]any{
				"login":    phone,
				"password": password,
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodPost, tt.body)
			authCtl.SignIn(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

func TestSignOut_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	authCtl := authController.New(useCases, setup.TestCfg, l)
	ctx := t.Context()

	// Seed user and session
	phone := "998903334455"
	u := domain.NewUser()
	u.ID = uuid.New()
	u.Username = stringPtr("signout_user")
	u.Phone = &phone
	u.IsApproved = true
	require.NoError(t, u.SetPassword("Str0ng!Pass"))
	require.NoError(t, repositories.Persistent.Postgres.User.Client.Create(ctx, u))

	sess := &domain.Session{
		ID:        uuid.New(),
		UserID:    u.ID,
		DeviceID:  uuid.New(),
		IPAddress: stringPtr("127.0.0.1"),
		UserAgent: stringPtr("test-agent"),
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}
	require.NoError(t, repositories.Persistent.Postgres.User.SessionRepo.Create(ctx, sess))

	tests := []struct {
		name          string
		body          map[string]any
		authenticated bool
		expectedCode  int
		checkResp     func(t *testing.T, body []byte)
	}{
		{
			name:          "success",
			body:          map[string]any{"session_id": sess.ID.String()},
			authenticated: true,
			expectedCode:  http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				// Verify session is revoked/deleted in DB
				_, err := repositories.Persistent.Postgres.User.SessionRepo.Get(ctx, &domain.SessionFilter{ID: &sess.ID})
				assert.Error(t, err)
			},
		},
		{
			name:          "unauthorized - no context",
			body:          map[string]any{"session_id": sess.ID.String()},
			authenticated: false,
			expectedCode:  http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodPost, tt.body)
			if tt.authenticated {
				setAuthContext(c, u.ID, sess)
			}

			authCtl.SignOut(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// User CRUD — Direct Controller Tests
// ---------------------------------------------------------------------------

func TestGetUser_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := clientController.New(useCases, setup.TestCfg, l)
	ctx := t.Context()

	// Seed user
	phone := "998905001001"
	u := domain.NewUser()
	u.ID = uuid.New()
	u.Username = stringPtr("get_test_user")
	u.Phone = &phone
	u.IsApproved = true
	require.NoError(t, u.SetPassword("Str0ng!Pass"))
	require.NoError(t, repositories.Persistent.Postgres.User.Client.Create(ctx, u))

	sess := seedSession(t, repositories, u.ID)

	tests := []struct {
		name         string
		userID       string
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name:         "success",
			userID:       u.ID.String(),
			expectedCode: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, "get_test_user", data["username"])
				assert.Equal(t, phone, data["phone"])
			},
		},
		{
			name:         "not found",
			userID:       uuid.New().String(),
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "invalid uuid",
			userID:       "not-a-uuid",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodGet, nil)
			c.Params = gin.Params{{Key: consts.ParamUserID, Value: tt.userID}}
			setAuthContext(c, u.ID, sess)

			controller.User(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

func TestListUsers_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := clientController.New(useCases, setup.TestCfg, l)
	ctx := t.Context()

	// Seed 3 users
	for i := range 3 {
		phone := "99890500100" + string(rune('1'+i))
		u := domain.NewUser()
		u.ID = uuid.New()
		u.Username = stringPtr("list_user_" + string(rune('a'+i)))
		u.Phone = &phone
		require.NoError(t, u.SetPassword("Str0ng!Pass"))
		require.NoError(t, repositories.Persistent.Postgres.User.Client.Create(ctx, u))
	}

	// Session for auth context
	ownerID := uuid.New()
	sess := &domain.Session{
		ID:        uuid.New(),
		UserID:    ownerID,
		DeviceID:  uuid.New(),
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}

	w, c := newGinContext(t, http.MethodGet, nil)
	setAuthContext(c, ownerID, sess)

	controller.Users(c)

	assert.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data, ok := resp["data"].([]any)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, len(data), 3)
}

func TestUpdateUser_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := clientController.New(useCases, setup.TestCfg, l)
	ctx := t.Context()

	// Seed user
	phone := "998906001001"
	u := domain.NewUser()
	u.ID = uuid.New()
	u.Username = stringPtr("update_user")
	u.Phone = &phone
	u.IsApproved = true
	require.NoError(t, u.SetPassword("Str0ng!Pass"))
	require.NoError(t, repositories.Persistent.Postgres.User.Client.Create(ctx, u))

	sess := seedSession(t, repositories, u.ID)

	tests := []struct {
		name         string
		userID       string
		body         map[string]any
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name:         "update username",
			userID:       u.ID.String(),
			body:         map[string]any{"username": "updated_name"},
			expectedCode: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				dbUser, err := repositories.Persistent.Postgres.User.Client.Get(ctx, &domain.UserFilter{ID: &u.ID})
				require.NoError(t, err)
				assert.Equal(t, "updated_name", *dbUser.Username)
			},
		},
		{
			name:         "update email",
			userID:       u.ID.String(),
			body:         map[string]any{"email": "newemail@example.com"},
			expectedCode: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				dbUser, err := repositories.Persistent.Postgres.User.Client.Get(ctx, &domain.UserFilter{ID: &u.ID})
				require.NoError(t, err)
				assert.Equal(t, "newemail@example.com", *dbUser.Email)
			},
		},
		{
			name:         "not found",
			userID:       uuid.New().String(),
			body:         map[string]any{"username": "ghost"},
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodPatch, tt.body)
			c.Params = gin.Params{{Key: consts.ParamUserID, Value: tt.userID}}
			setAuthContext(c, u.ID, sess)

			controller.Update(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

func TestDeleteUser_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := clientController.New(useCases, setup.TestCfg, l)
	ctx := t.Context()

	// Seed user
	phone := "998907001001"
	u := domain.NewUser()
	u.ID = uuid.New()
	u.Username = stringPtr("delete_user")
	u.Phone = &phone
	u.IsApproved = true
	require.NoError(t, u.SetPassword("Str0ng!Pass"))
	require.NoError(t, repositories.Persistent.Postgres.User.Client.Create(ctx, u))

	sess := seedSession(t, repositories, u.ID)

	tests := []struct {
		name         string
		userID       string
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name:         "success",
			userID:       u.ID.String(),
			expectedCode: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				_, err := repositories.Persistent.Postgres.User.Client.Get(ctx, &domain.UserFilter{ID: &u.ID})
				assert.Error(t, err)
			},
		},
		{
			name:         "already deleted - not found",
			userID:       u.ID.String(),
			expectedCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodDelete, nil)
			c.Params = gin.Params{{Key: consts.ParamUserID, Value: tt.userID}}
			setAuthContext(c, u.ID, sess)

			controller.Delete(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Comprehensive Flow — Sequential Multi-Step Test
// ---------------------------------------------------------------------------

func TestUser_ComprehensiveFlow_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	authCtl := authController.New(useCases, setup.TestCfg, l)
	controller := clientController.New(useCases, setup.TestCfg, l)

	type flowCtx struct {
		UserID    string
		SessionID string
		Token     string
	}
	fc := &flowCtx{}

	steps := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "Step 1: SignUp",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodPost, map[string]any{
					"username": "flow_user",
					"phone":    "998908001001",
					"password": "Str0ng!Pass",
					"session":  map[string]any{},
				})
				authCtl.SignUp(c)
				require.Equal(t, http.StatusCreated, w.Code, "body: %s", w.Body.String())

				var resp map[string]any
				json.Unmarshal(w.Body.Bytes(), &resp)
				data := resp["data"].(map[string]any)
				fc.UserID = data["user_id"].(string)
				fc.SessionID = data["session_id"].(string)
				fc.Token = data["access_token"].(string)
				assert.NotEmpty(t, fc.UserID)
			},
		},
		{
			name: "Step 2: SignIn",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodPost, map[string]any{
					"login":    "998908001001",
					"password": "Str0ng!Pass",
					"session":  map[string]any{},
				})
				authCtl.SignIn(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				var resp map[string]any
				json.Unmarshal(w.Body.Bytes(), &resp)
				data := resp["data"].(map[string]any)
				fc.Token = data["access_token"].(string)
				fc.SessionID = data["session_id"].(string)
			},
		},
		{
			name: "Step 3: Get Profile",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodGet, nil)
				c.Params = gin.Params{{Key: consts.ParamUserID, Value: fc.UserID}}
				sessID := uuid.MustParse(fc.SessionID)
				userID := uuid.MustParse(fc.UserID)
				c.Set(consts.CtxUserID, fc.UserID)
				c.Set(consts.CtxSessionID, sessID)
				c.Set(consts.CtxSession, &domain.Session{ID: sessID, UserID: userID})

				controller.User(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				var resp map[string]any
				json.Unmarshal(w.Body.Bytes(), &resp)
				data := resp["data"].(map[string]any)
				assert.Equal(t, "flow_user", data["username"])
			},
		},
		{
			name: "Step 4: Update Profile",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodPatch, map[string]any{"username": "flow_user_v2"})
				c.Params = gin.Params{{Key: consts.ParamUserID, Value: fc.UserID}}
				sessID := uuid.MustParse(fc.SessionID)
				userID := uuid.MustParse(fc.UserID)
				c.Set(consts.CtxUserID, fc.UserID)
				c.Set(consts.CtxSessionID, sessID)
				c.Set(consts.CtxSession, &domain.Session{ID: sessID, UserID: userID})

				controller.Update(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				// Verify in DB
				dbUser, err := repositories.Persistent.Postgres.User.Client.Get(
					t.Context(), &domain.UserFilter{ID: &userID},
				)
				require.NoError(t, err)
				assert.Equal(t, "flow_user_v2", *dbUser.Username)
			},
		},
		{
			name: "Step 5: Delete User",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodDelete, nil)
				c.Params = gin.Params{{Key: consts.ParamUserID, Value: fc.UserID}}
				sessID := uuid.MustParse(fc.SessionID)
				userID := uuid.MustParse(fc.UserID)
				c.Set(consts.CtxUserID, fc.UserID)
				c.Set(consts.CtxSessionID, sessID)
				c.Set(consts.CtxSession, &domain.Session{ID: sessID, UserID: userID})

				controller.Delete(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				// Verify deleted in DB
				_, err := repositories.Persistent.Postgres.User.Client.Get(
					t.Context(), &domain.UserFilter{ID: &userID},
				)
				assert.Error(t, err)
			},
		},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			step.run(t)
		})
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newGinContext(t *testing.T, method string, body any) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()

	var bodyReader *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewBuffer(jsonBody)
	} else {
		bodyReader = bytes.NewBuffer(nil)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, "/", bodyReader)
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return w, c
}

func setAuthContext(c *gin.Context, userID uuid.UUID, sess *domain.Session) {
	c.Set(consts.CtxUserID, userID.String())
	c.Set(consts.CtxSessionID, sess.ID)
	c.Set(consts.CtxSession, sess)
}

func seedSession(t *testing.T, repositories *repo.Repo, userID uuid.UUID) *domain.Session {
	t.Helper()
	sess := &domain.Session{
		ID:        uuid.New(),
		UserID:    userID,
		DeviceID:  uuid.New(),
		IPAddress: stringPtr("127.0.0.1"),
		UserAgent: stringPtr("test-agent"),
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}
	require.NoError(t, repositories.Persistent.Postgres.User.SessionRepo.Create(t.Context(), sess))
	return sess
}
