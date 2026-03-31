package client

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_ComprehensiveFlow(t *testing.T) {
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

		Token     string
		UserID    string
		SessionID string
	}

	ctx := &TestContext{
		Phone:    "+998901234567",
		Password: "P@ssw0rd!",
		Username: "testuser",
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
				ctx.SessionID = data["session_id"].(string)
				assert.NotEmpty(t, ctx.Token)
				assert.NotEmpty(t, ctx.UserID)
				assert.NotEmpty(t, ctx.SessionID)
			},
		},

		{
			name: "Step 3: Get Profile",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.Get(t, ctx.Token, ctx.UserID)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				data := body["data"].(map[string]any)
				assert.Equal(t, ctx.Phone, data["phone"])
				assert.Equal(t, ctx.Username, data["username"])
			},
		},
		{
			name: "Step 4: Update Profile",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				newName := "updated_name"
				resp := client.Update(t, ctx.Token, ctx.UserID, newName)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				// Verify Update
				respGet := client.Get(t, ctx.Token, ctx.UserID)
				defer respGet.Body.Close()

				var body map[string]any
				json.NewDecoder(respGet.Body).Decode(&body)
				data := body["data"].(map[string]any)
				assert.Equal(t, newName, data["username"])
			},
		},
		{
			name: "Step 5: Sign Out",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.SignOut(t, ctx.Token, ctx.UserID, ctx.SessionID)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)
			},
		},
		{
			name: "Step 6: Verify Unauthorized",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				// Allow some time for async token revocation if applicable
				time.Sleep(10 * time.Millisecond)
				resp := client.Get(t, ctx.Token, ctx.UserID)
				defer resp.Body.Close()
				assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
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
