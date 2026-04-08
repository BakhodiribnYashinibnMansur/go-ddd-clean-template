package content

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnnouncement_CRUDFlow(t *testing.T) {
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
		AnnouncementID string
	}

	ctx := &TestContext{
		Phone:    "+998901234701",
		Password: "P@ssw0rd!",
		Username: "announce_flow_user",
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
			name: "Step 3: Create Announcement",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				title := map[string]string{"uz": "Test sarlavha", "ru": "Тест заголовок", "en": "Test Title"}
				content := map[string]string{"uz": "Test matn", "ru": "Тест содержание", "en": "Test Content"}
				resp := client.CreateAnnouncement(t, ctx.Token, title, content, 1)
				defer resp.Body.Close()
				require.Equal(t, http.StatusCreated, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				data := body["data"].(map[string]any)
				ctx.AnnouncementID = data["id"].(string)
				assert.NotEmpty(t, ctx.AnnouncementID)
			},
		},
		{
			name: "Step 4: List Announcements",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.ListAnnouncements(t, ctx.Token)
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
			name: "Step 5: Get Announcement",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.GetAnnouncement(t, ctx.Token, ctx.AnnouncementID)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				data := body["data"].(map[string]any)
				assert.Equal(t, ctx.AnnouncementID, data["id"])
			},
		},
		{
			name: "Step 6: Update Announcement",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				updatedTitle := map[string]string{"uz": "Yangilangan sarlavha", "ru": "Обновлённый заголовок", "en": "Updated Title"}
				resp := client.UpdateAnnouncement(t, ctx.Token, ctx.AnnouncementID, map[string]any{
					"title":    updatedTitle,
					"priority": 2,
				})
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)
			},
		},
		{
			name: "Step 7: Verify Update",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.GetAnnouncement(t, ctx.Token, ctx.AnnouncementID)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				data := body["data"].(map[string]any)
				title := data["title"].(map[string]any)
				assert.Equal(t, "Updated Title", title["en"])
			},
		},
		{
			name: "Step 8: Delete Announcement",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.DeleteAnnouncement(t, ctx.Token, ctx.AnnouncementID)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)
			},
		},
		{
			name: "Step 9: Verify Deleted",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.GetAnnouncement(t, ctx.Token, ctx.AnnouncementID)
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
