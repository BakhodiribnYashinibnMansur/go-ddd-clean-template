package role

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/shared/domain/consts"
	roleController "gct/internal/controller/restapi/v1/authz/role"
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
// Role CRUD — Direct Controller Tests
// ---------------------------------------------------------------------------

func TestCreateRole_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := roleController.New(useCases, setup.TestCfg, l)

	tests := []struct {
		name         string
		body         map[string]any
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name:         "success",
			body:         map[string]any{"name": "test_role_create"},
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
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

func TestGetRole_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := roleController.New(useCases, setup.TestCfg, l)
	ctx := t.Context()

	// Seed a role
	role := &domain.Role{ID: uuid.New(), Name: "get_test_role", Description: stringPtr("test description")}
	require.NoError(t, repositories.Persistent.Postgres.Authz.Role.Create(ctx, role))

	tests := []struct {
		name         string
		roleID       string
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name:         "success",
			roleID:       role.ID.String(),
			expectedCode: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, "get_test_role", data["name"])
			},
		},
		{
			name:         "not found",
			roleID:       uuid.New().String(),
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "invalid uuid",
			roleID:       "not-a-uuid",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodGet, nil)
			c.Params = gin.Params{{Key: consts.ParamRoleID, Value: tt.roleID}}

			controller.Get(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

func TestListRoles_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := roleController.New(useCases, setup.TestCfg, l)
	ctx := t.Context()

	// Seed 3 roles
	for i := range 3 {
		role := &domain.Role{ID: uuid.New(), Name: "list_role_" + string(rune('a'+i))}
		require.NoError(t, repositories.Persistent.Postgres.Authz.Role.Create(ctx, role))
	}

	w, c := newGinContext(t, http.MethodGet, nil)
	// Set query params for pagination
	c.Request.URL.RawQuery = "limit=10&offset=0"

	controller.Gets(c)

	assert.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data, ok := resp["data"].([]any)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, len(data), 3)
}

func TestUpdateRole_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := roleController.New(useCases, setup.TestCfg, l)
	ctx := t.Context()

	// Seed a role
	role := &domain.Role{ID: uuid.New(), Name: "update_role"}
	require.NoError(t, repositories.Persistent.Postgres.Authz.Role.Create(ctx, role))

	tests := []struct {
		name         string
		roleID       string
		body         map[string]any
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name:         "update name",
			roleID:       role.ID.String(),
			body:         map[string]any{"name": "updated_role_name"},
			expectedCode: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				dbRole, err := repositories.Persistent.Postgres.Authz.Role.Get(ctx, &domain.RoleFilter{ID: &role.ID})
				require.NoError(t, err)
				assert.Equal(t, "updated_role_name", dbRole.Name)
			},
		},
		{
			name:         "not found",
			roleID:       uuid.New().String(),
			body:         map[string]any{"name": "ghost"},
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodPut, tt.body)
			c.Params = gin.Params{{Key: consts.ParamRoleID, Value: tt.roleID}}

			controller.Update(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

func TestDeleteRole_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := roleController.New(useCases, setup.TestCfg, l)
	ctx := t.Context()

	// Seed a role
	role := &domain.Role{ID: uuid.New(), Name: "delete_role"}
	require.NoError(t, repositories.Persistent.Postgres.Authz.Role.Create(ctx, role))

	tests := []struct {
		name         string
		roleID       string
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name:         "success",
			roleID:       role.ID.String(),
			expectedCode: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				_, err := repositories.Persistent.Postgres.Authz.Role.Get(ctx, &domain.RoleFilter{ID: &role.ID})
				assert.Error(t, err)
			},
		},
		{
			name:         "already deleted",
			roleID:       role.ID.String(),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodDelete, nil)
			c.Params = gin.Params{{Key: consts.ParamRoleID, Value: tt.roleID}}

			controller.Delete(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Role Permission Management — Direct Controller Tests
// ---------------------------------------------------------------------------

func TestAddPermission_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := roleController.New(useCases, setup.TestCfg, l)
	ctx := t.Context()

	// Seed role and permission
	role := &domain.Role{ID: uuid.New(), Name: "perm_add_role"}
	require.NoError(t, repositories.Persistent.Postgres.Authz.Role.Create(ctx, role))

	perm := &domain.Permission{ID: uuid.New(), Name: "perm_add_perm"}
	require.NoError(t, repositories.Persistent.Postgres.Authz.Permission.Create(ctx, perm))

	tests := []struct {
		name         string
		roleID       string
		permID       string
		expectedCode int
	}{
		{
			name:         "success",
			roleID:       role.ID.String(),
			permID:       perm.ID.String(),
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid role id",
			roleID:       "not-a-uuid",
			permID:       perm.ID.String(),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid perm id",
			roleID:       role.ID.String(),
			permID:       "not-a-uuid",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodPost, nil)
			c.Params = gin.Params{
				{Key: consts.ParamRoleID, Value: tt.roleID},
				{Key: consts.ParamPermID, Value: tt.permID},
			}

			controller.AddPermission(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestRemovePermission_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := roleController.New(useCases, setup.TestCfg, l)
	ctx := t.Context()

	// Seed role and permission, then add the permission to the role
	role := &domain.Role{ID: uuid.New(), Name: "perm_remove_role"}
	require.NoError(t, repositories.Persistent.Postgres.Authz.Role.Create(ctx, role))

	perm := &domain.Permission{ID: uuid.New(), Name: "perm_remove_perm"}
	require.NoError(t, repositories.Persistent.Postgres.Authz.Permission.Create(ctx, perm))

	require.NoError(t, repositories.Persistent.Postgres.Authz.Role.AddPermission(ctx, role.ID, perm.ID))

	tests := []struct {
		name         string
		roleID       string
		permID       string
		expectedCode int
	}{
		{
			name:         "success",
			roleID:       role.ID.String(),
			permID:       perm.ID.String(),
			expectedCode: http.StatusOK,
		},
		{
			name:         "invalid role id",
			roleID:       "not-a-uuid",
			permID:       perm.ID.String(),
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodDelete, nil)
			c.Params = gin.Params{
				{Key: consts.ParamRoleID, Value: tt.roleID},
				{Key: consts.ParamPermID, Value: tt.permID},
			}

			controller.RemovePermission(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
		})
	}
}

func TestGetPermissions_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	ctx := t.Context()

	// Seed role and permissions
	role := &domain.Role{ID: uuid.New(), Name: "get_perms_role"}
	require.NoError(t, repositories.Persistent.Postgres.Authz.Role.Create(ctx, role))

	perm1 := &domain.Permission{ID: uuid.New(), Name: "get_perms_perm_a"}
	perm2 := &domain.Permission{ID: uuid.New(), Name: "get_perms_perm_b"}
	require.NoError(t, repositories.Persistent.Postgres.Authz.Permission.Create(ctx, perm1))
	require.NoError(t, repositories.Persistent.Postgres.Authz.Permission.Create(ctx, perm2))

	require.NoError(t, repositories.Persistent.Postgres.Authz.Role.AddPermission(ctx, role.ID, perm1.ID))
	require.NoError(t, repositories.Persistent.Postgres.Authz.Role.AddPermission(ctx, role.ID, perm2.ID))

	// Verify via repo directly
	perms, err := repositories.Persistent.Postgres.Authz.Role.GetPermissions(ctx, role.ID)
	require.NoError(t, err)
	assert.Len(t, perms, 2)

	names := make(map[string]bool)
	for _, p := range perms {
		names[p.Name] = true
	}
	assert.True(t, names["get_perms_perm_a"])
	assert.True(t, names["get_perms_perm_b"])
}

// ---------------------------------------------------------------------------
// Comprehensive Flow — Sequential Multi-Step Test
// ---------------------------------------------------------------------------

func TestRole_ComprehensiveFlow_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	roleCtl := roleController.New(useCases, setup.TestCfg, l)
	permCtl := permController.New(useCases, setup.TestCfg, l)
	ctx := t.Context()

	type flowCtx struct {
		RoleID string
		PermID string
	}
	fc := &flowCtx{}

	steps := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "Step 1: Create Role",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodPost, map[string]any{
					"name":        "flow_role",
					"description": "integration test role",
				})
				roleCtl.Create(c)
				require.Equal(t, http.StatusCreated, w.Code, "body: %s", w.Body.String())

				// Retrieve the role by name to get its ID
				dbRole, err := repositories.Persistent.Postgres.Authz.Role.Get(ctx, &domain.RoleFilter{Name: stringPtr("flow_role")})
				require.NoError(t, err)
				fc.RoleID = dbRole.ID.String()
			},
		},
		{
			name: "Step 2: Get Role",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodGet, nil)
				c.Params = gin.Params{{Key: consts.ParamRoleID, Value: fc.RoleID}}

				roleCtl.Get(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				var resp map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, "flow_role", data["name"])
			},
		},
		{
			name: "Step 3: List Roles",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodGet, nil)
				c.Request.URL.RawQuery = "limit=10&offset=0"

				roleCtl.Gets(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				var resp map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				data, ok := resp["data"].([]any)
				assert.True(t, ok)
				assert.GreaterOrEqual(t, len(data), 1)
			},
		},
		{
			name: "Step 4: Update Role",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodPut, map[string]any{
					"name":        "flow_role_v2",
					"description": "updated description",
				})
				c.Params = gin.Params{{Key: consts.ParamRoleID, Value: fc.RoleID}}

				roleCtl.Update(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				// Verify in DB
				roleID := uuid.MustParse(fc.RoleID)
				dbRole, err := repositories.Persistent.Postgres.Authz.Role.Get(ctx, &domain.RoleFilter{ID: &roleID})
				require.NoError(t, err)
				assert.Equal(t, "flow_role_v2", dbRole.Name)
			},
		},
		{
			name: "Step 5: Create Permission and Add to Role",
			run: func(t *testing.T) {
				// Create a permission
				w, c := newGinContext(t, http.MethodPost, map[string]any{
					"name": "flow_permission",
				})
				permCtl.Create(c)
				require.Equal(t, http.StatusCreated, w.Code, "body: %s", w.Body.String())

				// Get the permission by name
				dbPerm, err := repositories.Persistent.Postgres.Authz.Permission.Get(ctx, &domain.PermissionFilter{Name: stringPtr("flow_permission")})
				require.NoError(t, err)
				fc.PermID = dbPerm.ID.String()

				// Add permission to role
				w2, c2 := newGinContext(t, http.MethodPost, nil)
				c2.Params = gin.Params{
					{Key: consts.ParamRoleID, Value: fc.RoleID},
					{Key: consts.ParamPermID, Value: fc.PermID},
				}
				roleCtl.AddPermission(c2)
				require.Equal(t, http.StatusOK, w2.Code, "body: %s", w2.Body.String())

				// Verify permission is linked
				roleID := uuid.MustParse(fc.RoleID)
				perms, err := repositories.Persistent.Postgres.Authz.Role.GetPermissions(ctx, roleID)
				require.NoError(t, err)
				assert.Len(t, perms, 1)
				assert.Equal(t, "flow_permission", perms[0].Name)
			},
		},
		{
			name: "Step 6: Remove Permission from Role",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodDelete, nil)
				c.Params = gin.Params{
					{Key: consts.ParamRoleID, Value: fc.RoleID},
					{Key: consts.ParamPermID, Value: fc.PermID},
				}

				roleCtl.RemovePermission(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				// Verify permission removed
				roleID := uuid.MustParse(fc.RoleID)
				perms, err := repositories.Persistent.Postgres.Authz.Role.GetPermissions(ctx, roleID)
				require.NoError(t, err)
				assert.Len(t, perms, 0)
			},
		},
		{
			name: "Step 7: Delete Role",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodDelete, nil)
				c.Params = gin.Params{{Key: consts.ParamRoleID, Value: fc.RoleID}}

				roleCtl.Delete(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				// Verify deleted
				roleID := uuid.MustParse(fc.RoleID)
				_, err := repositories.Persistent.Postgres.Authz.Role.Get(ctx, &domain.RoleFilter{ID: &roleID})
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
