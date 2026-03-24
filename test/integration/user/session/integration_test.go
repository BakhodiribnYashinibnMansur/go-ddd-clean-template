package session

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gct/internal/shared/domain/consts"
	sessionController "gct/internal/controller/restapi/v1/user/session"
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
// Session CRUD — Direct Controller Tests
// ---------------------------------------------------------------------------

func TestListSessions_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := sessionController.New(useCases, l)
	ctx := t.Context()

	// Seed user
	u := seedUser(t, repositories, "998910001001", "list_sessions_user")

	// Seed 3 sessions
	sessions := make([]*domain.Session, 3)
	for i := range 3 {
		sessions[i] = seedSession(t, repositories, u.ID)
	}

	tests := []struct {
		name          string
		authenticated bool
		expectedCode  int
		checkResp     func(t *testing.T, body []byte)
	}{
		{
			name:          "success - list all sessions",
			authenticated: true,
			expectedCode:  http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				data := resp["data"].([]any)
				assert.GreaterOrEqual(t, len(data), 3)

				// DB validation
				for _, sess := range data {
					s := sess.(map[string]any)
					sid := uuid.MustParse(s["id"].(string))
					dbSess, err := repositories.Persistent.Postgres.User.SessionRepo.Get(ctx, &domain.SessionFilter{ID: &sid})
					require.NoError(t, err)
					assert.Equal(t, u.ID, dbSess.UserID)
				}
			},
		},
		{
			name:          "unauthorized - no user context",
			authenticated: false,
			expectedCode:  http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodGet, nil)
			if tt.authenticated {
				setAuthContext(c, u.ID, sessions[0])
			}

			controller.Sessions(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

func TestGetSession_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := sessionController.New(useCases, l)

	u := seedUser(t, repositories, "998910002001", "get_session_user")
	sess := seedSession(t, repositories, u.ID)

	tests := []struct {
		name         string
		sessionID    string
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name:         "success",
			sessionID:    sess.ID.String(),
			expectedCode: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, sess.ID.String(), data["id"])
				assert.Equal(t, u.ID.String(), data["user_id"])
			},
		},
		{
			name:         "not found",
			sessionID:    uuid.New().String(),
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "invalid uuid",
			sessionID:    "invalid-uuid",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodGet, nil)
			c.Params = gin.Params{{Key: consts.ParamID, Value: tt.sessionID}}
			setAuthContext(c, u.ID, sess)

			controller.Session(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

func TestUpdateActivity_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := sessionController.New(useCases, l)
	ctx := t.Context()

	u := seedUser(t, repositories, "998910003001", "activity_user")
	sess := seedSession(t, repositories, u.ID)

	tests := []struct {
		name         string
		sessionID    string
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name:         "success",
			sessionID:    sess.ID.String(),
			expectedCode: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				dbSess, err := repositories.Persistent.Postgres.User.SessionRepo.Get(ctx, &domain.SessionFilter{ID: &sess.ID})
				require.NoError(t, err)
				assert.WithinDuration(t, time.Now(), dbSess.LastActivity, 5*time.Second)
				// Verify expiry was extended (~7 days from now)
				assert.WithinDuration(t, time.Now().Add(7*24*time.Hour), dbSess.ExpiresAt, 30*time.Second)
			},
		},
		{
			name:         "invalid uuid",
			sessionID:    "bad-uuid",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodPut, nil)
			c.Params = gin.Params{{Key: consts.ParamID, Value: tt.sessionID}}
			setAuthContext(c, u.ID, sess)

			controller.UpdateActivity(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

func TestDeleteSession_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := sessionController.New(useCases, l)
	ctx := t.Context()

	u := seedUser(t, repositories, "998910004001", "delete_session_user")
	sess := seedSession(t, repositories, u.ID)

	tests := []struct {
		name         string
		sessionID    string
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name:         "success",
			sessionID:    sess.ID.String(),
			expectedCode: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				_, err := repositories.Persistent.Postgres.User.SessionRepo.Get(ctx, &domain.SessionFilter{ID: &sess.ID})
				assert.Error(t, err, "session should be deleted/not found")
			},
		},
		{
			name:         "invalid uuid",
			sessionID:    "not-valid",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodDelete, nil)
			c.Params = gin.Params{{Key: consts.ParamID, Value: tt.sessionID}}
			setAuthContext(c, u.ID, sess)

			controller.Delete(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

func TestRevokeCurrent_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := sessionController.New(useCases, l)
	ctx := t.Context()

	u := seedUser(t, repositories, "998910005001", "revoke_current_user")
	sess := seedSession(t, repositories, u.ID)

	tests := []struct {
		name          string
		authenticated bool
		expectedCode  int
		checkResp     func(t *testing.T, body []byte)
	}{
		{
			name:          "success",
			authenticated: true,
			expectedCode:  http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				_, err := repositories.Persistent.Postgres.User.SessionRepo.Get(ctx, &domain.SessionFilter{ID: &sess.ID})
				assert.Error(t, err, "current session should be deleted")
			},
		},
		{
			name:          "unauthorized - no session in context",
			authenticated: false,
			expectedCode:  http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodDelete, nil)
			if tt.authenticated {
				setAuthContext(c, u.ID, sess)
			}

			controller.RevokeCurrent(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

func TestRevokeAll_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := sessionController.New(useCases, l)
	ctx := t.Context()

	u := seedUser(t, repositories, "998910006001", "revoke_all_user")

	// Create 3 sessions
	sessions := make([]*domain.Session, 3)
	for i := range 3 {
		sessions[i] = seedSession(t, repositories, u.ID)
	}

	tests := []struct {
		name          string
		authenticated bool
		body          map[string]any
		expectedCode  int
		checkResp     func(t *testing.T, body []byte)
	}{
		{
			name:          "success - revoke current session",
			authenticated: true,
			body:          map[string]any{"user_id": u.ID.String()},
			expectedCode:  http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				// The current implementation revokes only the current session
				dbSess, err := repositories.Persistent.Postgres.User.SessionRepo.Get(ctx, &domain.SessionFilter{ID: &sessions[0].ID})
				if err == nil {
					assert.True(t, dbSess.Revoked, "session should be revoked")
				}
			},
		},
		{
			name:          "unauthorized - no user context",
			authenticated: false,
			body:          map[string]any{},
			expectedCode:  http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodPost, tt.body)
			if tt.authenticated {
				setAuthContext(c, u.ID, sessions[0])
			}

			controller.RevokeAll(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

func TestCreateSession_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := sessionController.New(useCases, l)
	ctx := t.Context()

	u := seedUser(t, repositories, "998910007001", "create_session_user")
	authSess := seedSession(t, repositories, u.ID)

	deviceID := uuid.New()
	tests := []struct {
		name         string
		body         map[string]any
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name: "success",
			body: map[string]any{
				"user_id":    u.ID.String(),
				"device_id":  deviceID.String(),
				"ip_address": "192.168.1.100",
				"user_agent": "TestAgent/1.0",
			},
			expectedCode: http.StatusCreated,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				data := resp["data"].(map[string]any)
				sid := uuid.MustParse(data["id"].(string))

				dbSess, err := repositories.Persistent.Postgres.User.SessionRepo.Get(ctx, &domain.SessionFilter{ID: &sid})
				require.NoError(t, err)
				assert.Equal(t, u.ID, dbSess.UserID)
				assert.False(t, dbSess.Revoked)
			},
		},
		{
			name:         "empty body - creates session with defaults",
			body:         map[string]any{},
			expectedCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodPost, tt.body)
			setAuthContext(c, u.ID, authSess)

			controller.Create(c)

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

func TestSession_ComprehensiveFlow_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := sessionController.New(useCases, l)
	ctx := t.Context()

	u := seedUser(t, repositories, "998910008001", "flow_session_user")

	type flowCtx struct {
		Session1 *domain.Session
		Session2 *domain.Session
		Session3 *domain.Session
	}
	fc := &flowCtx{}

	steps := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "Step 1: Create 3 sessions",
			run: func(t *testing.T) {
				fc.Session1 = seedSession(t, repositories, u.ID)
				fc.Session2 = seedSession(t, repositories, u.ID)
				fc.Session3 = seedSession(t, repositories, u.ID)
			},
		},
		{
			name: "Step 2: List all sessions",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodGet, nil)
				setAuthContext(c, u.ID, fc.Session1)

				controller.Sessions(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				var resp map[string]any
				json.Unmarshal(w.Body.Bytes(), &resp)
				data := resp["data"].([]any)
				assert.GreaterOrEqual(t, len(data), 3)
			},
		},
		{
			name: "Step 3: Get session 2 by ID",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodGet, nil)
				c.Params = gin.Params{{Key: consts.ParamID, Value: fc.Session2.ID.String()}}
				setAuthContext(c, u.ID, fc.Session1)

				controller.Session(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				var resp map[string]any
				json.Unmarshal(w.Body.Bytes(), &resp)
				data := resp["data"].(map[string]any)
				assert.Equal(t, fc.Session2.ID.String(), data["id"])
			},
		},
		{
			name: "Step 4: Update activity on session 1",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodPut, nil)
				c.Params = gin.Params{{Key: consts.ParamID, Value: fc.Session1.ID.String()}}
				setAuthContext(c, u.ID, fc.Session1)

				controller.UpdateActivity(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				dbSess, err := repositories.Persistent.Postgres.User.SessionRepo.Get(ctx, &domain.SessionFilter{ID: &fc.Session1.ID})
				require.NoError(t, err)
				assert.WithinDuration(t, time.Now(), dbSess.LastActivity, 5*time.Second)
			},
		},
		{
			name: "Step 5: Delete session 3",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodDelete, nil)
				c.Params = gin.Params{{Key: consts.ParamID, Value: fc.Session3.ID.String()}}
				setAuthContext(c, u.ID, fc.Session1)

				controller.Delete(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				_, err := repositories.Persistent.Postgres.User.SessionRepo.Get(ctx, &domain.SessionFilter{ID: &fc.Session3.ID})
				assert.Error(t, err, "session 3 should be deleted")
			},
		},
		{
			name: "Step 6: Revoke current session (session 1)",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodDelete, nil)
				setAuthContext(c, u.ID, fc.Session1)

				controller.RevokeCurrent(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				_, err := repositories.Persistent.Postgres.User.SessionRepo.Get(ctx, &domain.SessionFilter{ID: &fc.Session1.ID})
				assert.Error(t, err, "session 1 should be deleted")
			},
		},
		{
			name: "Step 7: Verify only session 2 remains active",
			run: func(t *testing.T) {
				dbSess, err := repositories.Persistent.Postgres.User.SessionRepo.Get(ctx, &domain.SessionFilter{ID: &fc.Session2.ID})
				require.NoError(t, err)
				assert.False(t, dbSess.Revoked, "session 2 should still be active")
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

func stringPtr(s string) *string {
	return &s
}

func seedUser(t *testing.T, repositories *repo.Repo, phone, username string) *domain.User {
	t.Helper()
	u := domain.NewUser()
	u.ID = uuid.New()
	u.Username = stringPtr(username)
	u.Phone = &phone
	u.IsApproved = true
	require.NoError(t, u.SetPassword("Str0ng!Pass"))
	require.NoError(t, repositories.Persistent.Postgres.User.Client.Create(t.Context(), u))
	return u
}

func seedSession(t *testing.T, repositories *repo.Repo, userID uuid.UUID) *domain.Session {
	t.Helper()
	sess := &domain.Session{
		ID:        uuid.New(),
		UserID:    userID,
		DeviceID:  uuid.New(),
		IPAddress: stringPtr("127.0.0.1"),
		UserAgent: stringPtr("test-agent"),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		Revoked:   false,
	}
	require.NoError(t, repositories.Persistent.Postgres.User.SessionRepo.Create(t.Context(), sess))
	return sess
}
