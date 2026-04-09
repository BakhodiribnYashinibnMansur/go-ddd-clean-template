package admin

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSiteSetting_CRUDFlow(t *testing.T) {
	// 1. Setup Environment
	cleanDB(t)

	server := startTestServer()
	defer server.Close()

	client := New(server.URL)

	// 2. Shared State
	type TestContext struct {
		Phone    string
		Password string
		Username string

		Token        string
		UserID       string
		SiteSettingID string
	}

	ctx := &TestContext{
		Phone:    "+998901234567",
		Password: "P@ssw0rd!",
		Username: "admin_test_user",
	}

	// 3. Steps Table
	steps := []struct {
		name string
		run  func(t *testing.T, ctx *TestContext)
	}{
		{
			name: "Step 1: SignUp",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.SignUp(t, ctx.Username, ctx.Phone, ctx.Password)
				defer resp.Body.Close()
				require.Equal(t, http.StatusCreated, resp.StatusCode)
			},
		},
		{
			name: "Step 2: SignIn",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.SignIn(t, ctx.Phone, ctx.Password)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				data := body["data"].(map[string]any)
				ctx.Token = data["access_token"].(string)
				ctx.UserID = data["user_id"].(string)
				assert.NotEmpty(t, ctx.Token)
				assert.NotEmpty(t, ctx.UserID)
			},
		},
		{
			name: "Step 3: Create Site Setting",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.CreateSiteSetting(t, ctx.Token, map[string]any{
					"key":         "site_name",
					"value":       "My Application",
					"type":        "string",
					"description": "The name of the site",
				})
				defer resp.Body.Close()
				require.Equal(t, http.StatusCreated, resp.StatusCode)
			},
		},
		{
			name: "Step 4: List Site Settings",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.ListSiteSettings(t, ctx.Token)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				data := body["data"].([]any)
				require.GreaterOrEqual(t, len(data), 1)

				// Capture the ID from the first item for subsequent steps
				first := data[0].(map[string]any)
				ctx.SiteSettingID = first["id"].(string)
				assert.NotEmpty(t, ctx.SiteSettingID)
			},
		},
		{
			name: "Step 5: Get Site Setting",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.GetSiteSetting(t, ctx.Token, ctx.SiteSettingID)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				data := body["data"].(map[string]any)
				assert.Equal(t, "site_name", data["key"])
				assert.Equal(t, "My Application", data["value"])
			},
		},
		{
			name: "Step 6: Update Site Setting",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.UpdateSiteSetting(t, ctx.Token, ctx.SiteSettingID, map[string]any{
					"value": "Updated Application Name",
				})
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				// Verify update took effect
				getResp := client.GetSiteSetting(t, ctx.Token, ctx.SiteSettingID)
				defer getResp.Body.Close()
				require.Equal(t, http.StatusOK, getResp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(getResp.Body).Decode(&body)
				require.NoError(t, err)

				data := body["data"].(map[string]any)
				assert.Equal(t, "Updated Application Name", data["value"])
			},
		},
		{
			name: "Step 7: Delete Site Setting",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.DeleteSiteSetting(t, ctx.Token, ctx.SiteSettingID)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				// Verify deletion: GET should return 404
				getResp := client.GetSiteSetting(t, ctx.Token, ctx.SiteSettingID)
				defer getResp.Body.Close()
				assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
			},
		},
	}

	// 4. Execution
	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			step.run(t, ctx)
		})
	}
}
