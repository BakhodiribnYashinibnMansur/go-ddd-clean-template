package authz

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRoleFlow_CRUD exercises the full role lifecycle:
// sign up -> sign in -> create role -> list roles -> get role -> update role -> delete role.
func TestRoleFlow_CRUD(t *testing.T) {
	cleanDB(t)

	server := startTestServer()
	defer server.Close()

	c := New(server.URL)

	type TestContext struct {
		Phone    string
		Password string
		Username string

		Token  string
		RoleID string
	}

	ctx := &TestContext{
		Phone:    "+998901234700",
		Password: "P@ssw0rd!",
		Username: "authz_role_user",
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
			name: "Step 3: Create Role",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				desc := "Test role description"
				resp := c.CreateRole(t, ctx.Token, "test_admin_role", &desc)
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusCreated {
					body, _ := io.ReadAll(resp.Body)
					t.Fatalf("CreateRole: expected %d, got %d; body: %s", http.StatusCreated, resp.StatusCode, body)
				}

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, true, body["success"])
			},
		},
		{
			name: "Step 4: List Roles",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := c.ListRoles(t, ctx.Token)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				data, ok := body["data"].([]any)
				require.True(t, ok, "expected 'data' to be an array")
				require.GreaterOrEqual(t, len(data), 1, "expected at least one role")

				// Store the first role ID for subsequent steps
				firstRole := data[0].(map[string]any)
				ctx.RoleID = firstRole["id"].(string)
				assert.NotEmpty(t, ctx.RoleID)
			},
		},
		{
			name: "Step 5: Get Role",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := c.GetRole(t, ctx.Token, ctx.RoleID)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				data, ok := body["data"].(map[string]any)
				require.True(t, ok, "expected 'data' object in response")
				assert.Equal(t, ctx.RoleID, data["id"])
			},
		},
		{
			name: "Step 6: Update Role",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				newName := "updated_admin_role"
				newDesc := "Updated description"
				resp := c.UpdateRole(t, ctx.Token, ctx.RoleID, &newName, &newDesc)
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					body, _ := io.ReadAll(resp.Body)
					t.Fatalf("UpdateRole: expected %d, got %d; body: %s", http.StatusOK, resp.StatusCode, body)
				}

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, true, body["success"])

				// Verify update via GetRole
				respGet := c.GetRole(t, ctx.Token, ctx.RoleID)
				defer respGet.Body.Close()
				require.Equal(t, http.StatusOK, respGet.StatusCode)

				var getBody map[string]any
				json.NewDecoder(respGet.Body).Decode(&getBody)
				data := getBody["data"].(map[string]any)
				assert.Equal(t, newName, data["name"])
			},
		},
		{
			name: "Step 7: Delete Role",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := c.DeleteRole(t, ctx.Token, ctx.RoleID)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, true, body["success"])
			},
		},
		{
			name: "Step 8: Verify Role Deleted",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := c.GetRole(t, ctx.Token, ctx.RoleID)
				defer resp.Body.Close()
				assert.Equal(t, http.StatusNotFound, resp.StatusCode)
			},
		},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			step.run(t, ctx)
		})
	}
}
