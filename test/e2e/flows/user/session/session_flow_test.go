package session

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	userClient "gct/test/e2e/flows/user/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession_ComprehensiveFlow(t *testing.T) {
	// 1. Setup Environment
	cleanDB(t)

	server := startTestServer()
	defer server.Close()

	sClient := New(server.URL)
	uClient := userClient.New(server.URL)

	// 2. Shared State
	type TestContext struct {
		Phone    string
		Password string
		Username string

		Token1     string
		Token2     string
		SessionID2 string
	}

	ctx := &TestContext{
		Phone:    "998901112233",
		Password: "pass123",
		Username: "session_user",
	}

	// 3. Steps Table
	steps := []struct {
		name string
		run  func(t *testing.T, ctx *TestContext)
	}{
		{
			name: "Step 1: SignUp",
			run: func(t *testing.T, ctx *TestContext) {
				resp := uClient.SignUp(t, ctx.Username, ctx.Phone, ctx.Password)
				defer resp.Body.Close()
				require.Equal(t, http.StatusCreated, resp.StatusCode)
			},
		},
		{
			name: "Step 2: Login Device 1",
			run: func(t *testing.T, ctx *TestContext) {
				resp := uClient.SignIn(t, ctx.Phone, ctx.Password)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)
				data := body["data"].(map[string]any)
				ctx.Token1 = data["access_token"].(string)
				assert.NotEmpty(t, ctx.Token1)
			},
		},
		{
			name: "Step 3: Login Device 2",
			run: func(t *testing.T, ctx *TestContext) {
				resp := uClient.SignIn(t, ctx.Phone, ctx.Password)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)
				data := body["data"].(map[string]any)
				ctx.Token2 = data["access_token"].(string)
				assert.NotEmpty(t, ctx.Token2)

				if sid, ok := data["session_id"].(string); ok {
					ctx.SessionID2 = sid
				}
			},
		},
		{
			name: "Step 4: List Sessions",
			run: func(t *testing.T, ctx *TestContext) {
				resp := sClient.List(t, ctx.Token1)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				data := body["data"].([]any)
				assert.GreaterOrEqual(t, len(data), 2)
			},
		},
		{
			name: "Step 5: Revoke Session 2",
			run: func(t *testing.T, ctx *TestContext) {
				if ctx.SessionID2 != "" {
					resp := sClient.Delete(t, ctx.Token1, ctx.SessionID2)
					defer resp.Body.Close()
					assert.Equal(t, http.StatusOK, resp.StatusCode)
				} else {
					t.Log("Skipping revoke specific session test as ID not captured")
				}
			},
		},
		{
			name: "Step 6: Revoke All",
			run: func(t *testing.T, ctx *TestContext) {
				resp := sClient.RevokeAll(t, ctx.Token1)
				defer resp.Body.Close()
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			},
		},
		{
			name: "Step 7: Verify Token1 Invalid",
			run: func(t *testing.T, ctx *TestContext) {
				time.Sleep(10 * time.Millisecond)
				resp := sClient.List(t, ctx.Token1)
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
