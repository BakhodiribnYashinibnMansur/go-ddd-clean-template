package authz

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPermissionFlow_CRUD exercises the full permission lifecycle:
// sign up -> sign in -> create permission -> create scope -> assign scope to permission
// -> list permissions -> delete permission.
func TestPermissionFlow_CRUD(t *testing.T) {
	cleanDB(t)

	server := startTestServer()
	defer server.Close()

	c := New(server.URL)

	type TestContext struct {
		Phone    string
		Password string
		Username string

		Token        string
		PermissionID string
	}

	ctx := &TestContext{
		Phone:    "+998901234701",
		Password: "P@ssw0rd!",
		Username: "authz_perm_user",
	}

	steps := []struct {
		name string
		run  func(t *testing.T, ctx *TestContext)
	}{
		{
			name: "Step 1: SignUp",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := c.SignUp(t, ctx.Username, ctx.Phone, ctx.Password)
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusCreated {
					body, _ := io.ReadAll(resp.Body)
					t.Fatalf("SignUp: expected %d, got %d; body: %s", http.StatusCreated, resp.StatusCode, body)
				}
			},
		},
		{
			name: "Step 2: SignIn",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := c.SignIn(t, ctx.Phone, ctx.Password)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				data := body["data"].(map[string]any)
				ctx.Token = data["access_token"].(string)
				assert.NotEmpty(t, ctx.Token)
			},
		},
		{
			name: "Step 3: Create Permission",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				desc := "Permission for managing users"
				resp := c.CreatePermission(t, ctx.Token, "users.manage", &desc)
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusCreated {
					body, _ := io.ReadAll(resp.Body)
					t.Fatalf("CreatePermission: expected %d, got %d; body: %s", http.StatusCreated, resp.StatusCode, body)
				}

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, true, body["success"])
			},
		},
		{
			name: "Step 4: List Permissions to get ID",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := c.ListPermissions(t, ctx.Token)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				data, ok := body["data"].([]any)
				require.True(t, ok, "expected 'data' to be an array")
				require.GreaterOrEqual(t, len(data), 1, "expected at least one permission")

				// Find the permission we just created
				for _, item := range data {
					perm := item.(map[string]any)
					if perm["name"] == "users.manage" {
						ctx.PermissionID = perm["id"].(string)
						break
					}
				}
				require.NotEmpty(t, ctx.PermissionID, "could not find created permission")
			},
		},
		{
			name: "Step 5: Create Scope",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := c.CreateScope(t, ctx.Token, "/api/v1/users", "GET")
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusCreated {
					body, _ := io.ReadAll(resp.Body)
					t.Fatalf("CreateScope: expected %d, got %d; body: %s", http.StatusCreated, resp.StatusCode, body)
				}

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, true, body["success"])
			},
		},
		{
			name: "Step 6: Assign Scope to Permission",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := c.AssignScopeToPermission(t, ctx.Token, ctx.PermissionID, "/api/v1/users", "GET")
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					body, _ := io.ReadAll(resp.Body)
					t.Fatalf("AssignScope: expected %d, got %d; body: %s", http.StatusOK, resp.StatusCode, body)
				}

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, true, body["success"])
			},
		},
		{
			name: "Step 7: List Scopes",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := c.ListScopes(t, ctx.Token)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				data, ok := body["data"].([]any)
				require.True(t, ok, "expected 'data' to be an array")
				require.GreaterOrEqual(t, len(data), 1, "expected at least one scope")
			},
		},
		{
			name: "Step 8: Delete Permission",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := c.DeletePermission(t, ctx.Token, ctx.PermissionID)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, true, body["success"])
			},
		},
		{
			name: "Step 9: Verify Permission Deleted",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := c.ListPermissions(t, ctx.Token)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				json.NewDecoder(resp.Body).Decode(&body)

				data, ok := body["data"].([]any)
				if ok {
					for _, item := range data {
						perm := item.(map[string]any)
						assert.NotEqual(t, ctx.PermissionID, perm["id"],
							"deleted permission should not appear in list")
					}
				}
			},
		},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			step.run(t, ctx)
		})
	}
}
