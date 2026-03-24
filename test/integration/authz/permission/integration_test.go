package permission

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/shared/domain/consts"
	permController "gct/internal/controller/restapi/v1/authz/permission"
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
// Permission CRUD — Direct Controller Tests
// ---------------------------------------------------------------------------

func TestCreatePermission_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := permController.New(useCases, setup.TestCfg, l)

	tests := []struct {
		name         string
		body         map[string]any
		expectedCode int
	}{
		{
			name:         "success",
			body:         map[string]any{"name": "test_perm_create"},
			expectedCode: http.StatusCreated,
		},
		{
			name:         "bad request - empty body",
			body:         map[string]any{},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodPost, tt.body)
			controller.Create(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestGetPermission_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := permController.New(useCases, setup.TestCfg, l)
	ctx := t.Context()

	// Seed a permission
	perm := &domain.Permission{ID: uuid.New(), Name: "get_test_perm", Description: stringPtr("test description")}
	require.NoError(t, repositories.Persistent.Postgres.Authz.Permission.Create(ctx, perm))

	tests := []struct {
		name         string
		permID       string
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name:         "success",
			permID:       perm.ID.String(),
			expectedCode: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, "get_test_perm", data["name"])
			},
		},
		{
			name:         "not found",
			permID:       uuid.New().String(),
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "invalid uuid",
			permID:       "not-a-uuid",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodGet, nil)
			c.Params = gin.Params{{Key: consts.ParamPermID, Value: tt.permID}}

			controller.Get(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

func TestListPermissions_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := permController.New(useCases, setup.TestCfg, l)
	ctx := t.Context()

	// Seed 3 permissions
	for i := range 3 {
		perm := &domain.Permission{ID: uuid.New(), Name: "list_perm_" + string(rune('a'+i))}
		require.NoError(t, repositories.Persistent.Postgres.Authz.Permission.Create(ctx, perm))
	}

	w, c := newGinContext(t, http.MethodGet, nil)
	c.Request.URL.RawQuery = "limit=10&offset=0"

	controller.Gets(c)

	assert.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data, ok := resp["data"].([]any)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, len(data), 3)
}

func TestUpdatePermission_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := permController.New(useCases, setup.TestCfg, l)
	ctx := t.Context()

	// Seed a permission
	perm := &domain.Permission{ID: uuid.New(), Name: "update_perm"}
	require.NoError(t, repositories.Persistent.Postgres.Authz.Permission.Create(ctx, perm))

	tests := []struct {
		name         string
		permID       string
		body         map[string]any
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name:         "update name",
			permID:       perm.ID.String(),
			body:         map[string]any{"name": "updated_perm_name"},
			expectedCode: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				dbPerm, err := repositories.Persistent.Postgres.Authz.Permission.Get(ctx, &domain.PermissionFilter{ID: &perm.ID})
				require.NoError(t, err)
				assert.Equal(t, "updated_perm_name", dbPerm.Name)
			},
		},
		{
			name:         "not found",
			permID:       uuid.New().String(),
			body:         map[string]any{"name": "ghost"},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodPut, tt.body)
			c.Params = gin.Params{{Key: consts.ParamPermID, Value: tt.permID}}

			controller.Update(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

func TestDeletePermission_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := permController.New(useCases, setup.TestCfg, l)
	ctx := t.Context()

	// Seed a permission
	perm := &domain.Permission{ID: uuid.New(), Name: "delete_perm"}
	require.NoError(t, repositories.Persistent.Postgres.Authz.Permission.Create(ctx, perm))

	tests := []struct {
		name         string
		permID       string
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name:         "success",
			permID:       perm.ID.String(),
			expectedCode: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				_, err := repositories.Persistent.Postgres.Authz.Permission.Get(ctx, &domain.PermissionFilter{ID: &perm.ID})
				assert.Error(t, err)
			},
		},
		{
			name:         "already deleted",
			permID:       perm.ID.String(),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodDelete, nil)
			c.Params = gin.Params{{Key: consts.ParamPermID, Value: tt.permID}}

			controller.Delete(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Scope Management — Direct Controller Tests
// ---------------------------------------------------------------------------

func TestAssignScope_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := permController.New(useCases, setup.TestCfg, l)
	ctx := t.Context()

	// Seed a permission
	perm := &domain.Permission{ID: uuid.New(), Name: "scope_assign_perm"}
	require.NoError(t, repositories.Persistent.Postgres.Authz.Permission.Create(ctx, perm))

	tests := []struct {
		name         string
		permID       string
		body         map[string]any
		expectedCode int
	}{
		{
			name:   "success",
			permID: perm.ID.String(),
			body: map[string]any{
				"path":   "/api/v1/test",
				"method": "GET",
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid perm id",
			permID:       "not-a-uuid",
			body:         map[string]any{"path": "/api/v1/test", "method": "GET"},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "bad request - missing fields",
			permID:       perm.ID.String(),
			body:         map[string]any{},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodPost, tt.body)
			c.Params = gin.Params{{Key: consts.ParamPermID, Value: tt.permID}}

			controller.AssignScope(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestRemoveScope_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := permController.New(useCases, setup.TestCfg, l)
	ctx := t.Context()

	// Seed a permission and assign a scope
	perm := &domain.Permission{ID: uuid.New(), Name: "scope_remove_perm"}
	require.NoError(t, repositories.Persistent.Postgres.Authz.Permission.Create(ctx, perm))
	require.NoError(t, repositories.Persistent.Postgres.Authz.Permission.AddScope(ctx, perm.ID, "/api/v1/remove-test", "DELETE"))

	tests := []struct {
		name         string
		permID       string
		body         map[string]any
		expectedCode int
	}{
		{
			name:   "success",
			permID: perm.ID.String(),
			body: map[string]any{
				"path":   "/api/v1/remove-test",
				"method": "DELETE",
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid perm id",
			permID:       "not-a-uuid",
			body:         map[string]any{"path": "/api/v1/remove-test", "method": "DELETE"},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodDelete, tt.body)
			c.Params = gin.Params{{Key: consts.ParamPermID, Value: tt.permID}}

			controller.RemoveScope(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestGetScopes_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	ctx := t.Context()

	// Seed a permission and assign scopes
	perm := &domain.Permission{ID: uuid.New(), Name: "get_scopes_perm"}
	require.NoError(t, repositories.Persistent.Postgres.Authz.Permission.Create(ctx, perm))

	require.NoError(t, repositories.Persistent.Postgres.Authz.Permission.AddScope(ctx, perm.ID, "/api/v1/scope-a", "GET"))
	require.NoError(t, repositories.Persistent.Postgres.Authz.Permission.AddScope(ctx, perm.ID, "/api/v1/scope-b", "POST"))

	// Verify via repo directly
	scopes, err := repositories.Persistent.Postgres.Authz.Permission.GetScopes(ctx, perm.ID)
	require.NoError(t, err)
	assert.Len(t, scopes, 2)

	paths := make(map[string]string)
	for _, s := range scopes {
		paths[s.Path] = s.Method
	}
	assert.Equal(t, "GET", paths["/api/v1/scope-a"])
	assert.Equal(t, "POST", paths["/api/v1/scope-b"])
}

// ---------------------------------------------------------------------------
// Comprehensive Flow — Sequential Multi-Step Test
// ---------------------------------------------------------------------------

func TestPermission_ComprehensiveFlow_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := permController.New(useCases, setup.TestCfg, l)
	ctx := t.Context()

	type flowCtx struct {
		PermID string
	}
	fc := &flowCtx{}

	steps := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "Step 1: Create Permission",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodPost, map[string]any{
					"name":        "flow_permission",
					"description": "integration test permission",
				})
				controller.Create(c)
				require.Equal(t, http.StatusCreated, w.Code, "body: %s", w.Body.String())

				// Retrieve by name to get ID
				dbPerm, err := repositories.Persistent.Postgres.Authz.Permission.Get(ctx, &domain.PermissionFilter{Name: stringPtr("flow_permission")})
				require.NoError(t, err)
				fc.PermID = dbPerm.ID.String()
			},
		},
		{
			name: "Step 2: Get Permission",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodGet, nil)
				c.Params = gin.Params{{Key: consts.ParamPermID, Value: fc.PermID}}

				controller.Get(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				var resp map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, "flow_permission", data["name"])
			},
		},
		{
			name: "Step 3: List Permissions",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodGet, nil)
				c.Request.URL.RawQuery = "limit=10&offset=0"

				controller.Gets(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				var resp map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				data, ok := resp["data"].([]any)
				assert.True(t, ok)
				assert.GreaterOrEqual(t, len(data), 1)
			},
		},
		{
			name: "Step 4: Update Permission",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodPut, map[string]any{
					"name":        "flow_permission_v2",
					"description": "updated description",
				})
				c.Params = gin.Params{{Key: consts.ParamPermID, Value: fc.PermID}}

				controller.Update(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				// Verify in DB
				permID := uuid.MustParse(fc.PermID)
				dbPerm, err := repositories.Persistent.Postgres.Authz.Permission.Get(ctx, &domain.PermissionFilter{ID: &permID})
				require.NoError(t, err)
				assert.Equal(t, "flow_permission_v2", dbPerm.Name)
			},
		},
		{
			name: "Step 5: Assign Scope",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodPost, map[string]any{
					"path":   "/api/v1/flow-test",
					"method": "GET",
				})
				c.Params = gin.Params{{Key: consts.ParamPermID, Value: fc.PermID}}

				controller.AssignScope(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				// Verify scope is linked
				permID := uuid.MustParse(fc.PermID)
				scopes, err := repositories.Persistent.Postgres.Authz.Permission.GetScopes(ctx, permID)
				require.NoError(t, err)
				assert.Len(t, scopes, 1)
				assert.Equal(t, "/api/v1/flow-test", scopes[0].Path)
				assert.Equal(t, "GET", scopes[0].Method)
			},
		},
		{
			name: "Step 6: Remove Scope",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodDelete, map[string]any{
					"path":   "/api/v1/flow-test",
					"method": "GET",
				})
				c.Params = gin.Params{{Key: consts.ParamPermID, Value: fc.PermID}}

				controller.RemoveScope(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				// Verify scope removed
				permID := uuid.MustParse(fc.PermID)
				scopes, err := repositories.Persistent.Postgres.Authz.Permission.GetScopes(ctx, permID)
				require.NoError(t, err)
				assert.Len(t, scopes, 0)
			},
		},
		{
			name: "Step 7: Delete Permission",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodDelete, nil)
				c.Params = gin.Params{{Key: consts.ParamPermID, Value: fc.PermID}}

				controller.Delete(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				// Verify deleted
				permID := uuid.MustParse(fc.PermID)
				_, err := repositories.Persistent.Postgres.Authz.Permission.Get(ctx, &domain.PermissionFilter{ID: &permID})
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
