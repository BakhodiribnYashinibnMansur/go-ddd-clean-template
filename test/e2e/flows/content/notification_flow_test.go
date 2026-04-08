package content

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotification_CRUDFlow(t *testing.T) {
	cleanDB(t)

	server := startTestServer()
	defer server.Close()

	client := New(server.URL)

	type TestContext struct {
		Phone    string
		Password string
		Username string

		Token          string
		UserID         string
		SessionID      string
		NotificationID string
	}

	ctx := &TestContext{
		Phone:    "+998901234700",
		Password: "P@ssw0rd!",
		Username: "notif_flow_user",
	}

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
			},
		},
		{
			name: "Step 3: Create Notification",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.CreateNotification(t, ctx.Token, ctx.UserID, "Test Title", "Test message body", "info")
				defer resp.Body.Close()
				require.Equal(t, http.StatusCreated, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				data := body["data"].(map[string]any)
				ctx.NotificationID = data["id"].(string)
				assert.NotEmpty(t, ctx.NotificationID)
			},
		},
		{
			name: "Step 4: List Notifications",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.ListNotifications(t, ctx.Token)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				data, ok := body["data"].([]any)
				require.True(t, ok, "expected data to be an array")
				assert.GreaterOrEqual(t, len(data), 1)
			},
		},
		{
			name: "Step 5: Get Notification",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.GetNotification(t, ctx.Token, ctx.NotificationID)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				data := body["data"].(map[string]any)
				assert.Equal(t, ctx.NotificationID, data["id"])
				assert.Equal(t, "Test Title", data["title"])
			},
		},
		{
			name: "Step 6: Delete Notification",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.DeleteNotification(t, ctx.Token, ctx.NotificationID)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)
			},
		},
		{
			name: "Step 7: Verify Deleted",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.GetNotification(t, ctx.Token, ctx.NotificationID)
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
