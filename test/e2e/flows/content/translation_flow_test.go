package content

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslation_CRUDFlow(t *testing.T) {
	cleanDB(t)

	server := startTestServer()
	defer server.Close()

	client := New(server.URL)

	type TestContext struct {
		Phone    string
		Password string
		Username string

		Token         string
		UserID        string
		SessionID     string
		TranslationID string
	}

	ctx := &TestContext{
		Phone:    "+998901234702",
		Password: "P@ssw0rd!",
		Username: "transl_flow_user",
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
			name: "Step 3: Create Translation",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.CreateTranslation(t, ctx.Token, "greeting.hello", "en", "Hello", "common")
				defer resp.Body.Close()
				require.Equal(t, http.StatusCreated, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				data := body["data"].(map[string]any)
				ctx.TranslationID = data["id"].(string)
				assert.NotEmpty(t, ctx.TranslationID)
			},
		},
		{
			name: "Step 4: List Translations",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.ListTranslations(t, ctx.Token)
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
			name: "Step 5: Get Translation",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.GetTranslation(t, ctx.Token, ctx.TranslationID)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				data := body["data"].(map[string]any)
				assert.Equal(t, ctx.TranslationID, data["id"])
				assert.Equal(t, "Hello", data["value"])
			},
		},
		{
			name: "Step 6: Update Translation",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.UpdateTranslation(t, ctx.Token, ctx.TranslationID, map[string]any{
					"value": "Hello, World!",
				})
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)
			},
		},
		{
			name: "Step 7: Verify Update",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.GetTranslation(t, ctx.Token, ctx.TranslationID)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)

				var body map[string]any
				err := json.NewDecoder(resp.Body).Decode(&body)
				require.NoError(t, err)

				data := body["data"].(map[string]any)
				assert.Equal(t, "Hello, World!", data["value"])
			},
		},
		{
			name: "Step 8: Delete Translation",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.DeleteTranslation(t, ctx.Token, ctx.TranslationID)
				defer resp.Body.Close()
				require.Equal(t, http.StatusOK, resp.StatusCode)
			},
		},
		{
			name: "Step 9: Verify Deleted",
			run: func(t *testing.T, ctx *TestContext) {
				t.Helper()
				resp := client.GetTranslation(t, ctx.Token, ctx.TranslationID)
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
