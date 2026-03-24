package webhook

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/internal/domain"
	webhookController "gct/internal/controller/restapi/v1/webhook"
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

func newGinContextWithQuery(t *testing.T, method, query string, body any) (*httptest.ResponseRecorder, *gin.Context) {
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
	req := httptest.NewRequest(method, "/?"+query, bodyReader)
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return w, c
}

func stringPtr(s string) *string { return &s }

func createTestWebhook(name, url string) domain.CreateWebhookRequest {
	return domain.CreateWebhookRequest{
		Name:     name,
		URL:      url,
		Secret:   "test-secret",
		Events:   []string{"user.created", "user.updated"},
		Headers:  map[string]any{"X-Custom": "value"},
		IsActive: true,
	}
}

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestCreate_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := webhookController.New(useCases.Webhook, setup.TestCfg, l)

	tests := []struct {
		name         string
		body         map[string]any
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name: "success",
			body: map[string]any{
				"name":      "Test Webhook",
				"url":       "https://example.com/hook",
				"secret":    "s3cret",
				"events":    []string{"user.created"},
				"headers":   map[string]any{"X-Key": "val"},
				"is_active": true,
			},
			expectedCode: http.StatusCreated,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, "Test Webhook", data["name"])
				assert.Equal(t, "https://example.com/hook", data["url"])
				assert.NotEmpty(t, data["id"])
			},
		},
		{
			name: "bad request - missing name",
			body: map[string]any{
				"url": "https://example.com/hook",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "bad request - invalid url",
			body: map[string]any{
				"name": "Bad URL Webhook",
				"url":  "not-a-url",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "bad request - missing url",
			body: map[string]any{
				"name": "No URL",
			},
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

// ---------------------------------------------------------------------------
// Get
// ---------------------------------------------------------------------------

func TestGet_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := webhookController.New(useCases.Webhook, setup.TestCfg, l)
	ctx := t.Context()

	created, err := useCases.Webhook.Create(ctx, createTestWebhook("Get Webhook", "https://example.com/get"))
	require.NoError(t, err)

	tests := []struct {
		name         string
		id           string
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name:         "success",
			id:           created.ID.String(),
			expectedCode: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, "Get Webhook", data["name"])
			},
		},
		{
			name:         "not found",
			id:           uuid.New().String(),
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "invalid uuid",
			id:           "not-a-uuid",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodGet, nil)
			c.Params = gin.Params{{Key: "id", Value: tt.id}}

			controller.Get(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

func TestList_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := webhookController.New(useCases.Webhook, setup.TestCfg, l)
	ctx := t.Context()

	// Seed 3 webhooks
	for i := range 3 {
		_, err := useCases.Webhook.Create(ctx, createTestWebhook(
			"List Webhook "+string(rune('A'+i)),
			"https://example.com/list/"+string(rune('a'+i)),
		))
		require.NoError(t, err)
	}

	t.Run("list all", func(t *testing.T) {
		w, c := newGinContextWithQuery(t, http.MethodGet, "limit=10&offset=0", nil)
		controller.List(c)

		assert.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		data, ok := resp["data"].([]any)
		assert.True(t, ok)
		assert.GreaterOrEqual(t, len(data), 3)
	})

	t.Run("pagination - limit 1", func(t *testing.T) {
		w, c := newGinContextWithQuery(t, http.MethodGet, "limit=1&offset=0", nil)
		controller.List(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		data := resp["data"].([]any)
		assert.Len(t, data, 1)
	})

	t.Run("search by name", func(t *testing.T) {
		w, c := newGinContextWithQuery(t, http.MethodGet, "limit=10&offset=0&search=List+Webhook", nil)
		controller.List(c)

		assert.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		data := resp["data"].([]any)
		assert.GreaterOrEqual(t, len(data), 1)
	})
}

// ---------------------------------------------------------------------------
// Update
// ---------------------------------------------------------------------------

func TestUpdate_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := webhookController.New(useCases.Webhook, setup.TestCfg, l)
	ctx := t.Context()

	created, err := useCases.Webhook.Create(ctx, createTestWebhook("Update Webhook", "https://example.com/update"))
	require.NoError(t, err)

	tests := []struct {
		name         string
		id           string
		body         map[string]any
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name:         "update name",
			id:           created.ID.String(),
			body:         map[string]any{"name": "Updated Webhook Name"},
			expectedCode: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, "Updated Webhook Name", data["name"])
			},
		},
		{
			name:         "update url",
			id:           created.ID.String(),
			body:         map[string]any{"url": "https://example.com/new-url"},
			expectedCode: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, "https://example.com/new-url", data["url"])
			},
		},
		{
			name:         "not found",
			id:           uuid.New().String(),
			body:         map[string]any{"name": "Ghost"},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "invalid uuid",
			id:           "not-a-uuid",
			body:         map[string]any{"name": "Bad"},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodPatch, tt.body)
			c.Params = gin.Params{{Key: "id", Value: tt.id}}

			controller.Update(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Delete
// ---------------------------------------------------------------------------

func TestDelete_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := webhookController.New(useCases.Webhook, setup.TestCfg, l)
	ctx := t.Context()

	created, err := useCases.Webhook.Create(ctx, createTestWebhook("Delete Webhook", "https://example.com/delete"))
	require.NoError(t, err)

	tests := []struct {
		name         string
		id           string
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name:         "success",
			id:           created.ID.String(),
			expectedCode: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				_, err := useCases.Webhook.GetByID(ctx, created.ID)
				assert.Error(t, err)
			},
		},
		{
			name:         "already deleted - error",
			id:           created.ID.String(),
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "invalid uuid",
			id:           "not-a-uuid",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, c := newGinContext(t, http.MethodDelete, nil)
			c.Params = gin.Params{{Key: "id", Value: tt.id}}

			controller.Delete(c)

			assert.Equal(t, tt.expectedCode, w.Code, "body: %s", w.Body.String())
			if tt.checkResp != nil {
				tt.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Comprehensive Flow
// ---------------------------------------------------------------------------

func TestWebhook_ComprehensiveFlow(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := webhookController.New(useCases.Webhook, setup.TestCfg, l)

	var webhookID string

	steps := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "Step 1: Create",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodPost, map[string]any{
					"name":      "Flow Webhook",
					"url":       "https://example.com/flow",
					"secret":    "flow-secret",
					"events":    []string{"order.placed"},
					"is_active": true,
				})
				controller.Create(c)
				require.Equal(t, http.StatusCreated, w.Code, "body: %s", w.Body.String())

				var resp map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				data := resp["data"].(map[string]any)
				webhookID = data["id"].(string)
				assert.NotEmpty(t, webhookID)
			},
		},
		{
			name: "Step 2: Get",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodGet, nil)
				c.Params = gin.Params{{Key: "id", Value: webhookID}}

				controller.Get(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				var resp map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, "Flow Webhook", data["name"])
			},
		},
		{
			name: "Step 3: List",
			run: func(t *testing.T) {
				w, c := newGinContextWithQuery(t, http.MethodGet, "limit=10&offset=0", nil)
				controller.List(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				var resp map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				data := resp["data"].([]any)
				assert.GreaterOrEqual(t, len(data), 1)
			},
		},
		{
			name: "Step 4: Update",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodPatch, map[string]any{"name": "Updated Flow Webhook"})
				c.Params = gin.Params{{Key: "id", Value: webhookID}}

				controller.Update(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				var resp map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, "Updated Flow Webhook", data["name"])
			},
		},
		{
			name: "Step 5: Delete",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodDelete, nil)
				c.Params = gin.Params{{Key: "id", Value: webhookID}}

				controller.Delete(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())
			},
		},
		{
			name: "Step 6: Verify Deleted",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodGet, nil)
				c.Params = gin.Params{{Key: "id", Value: webhookID}}

				controller.Get(c)
				assert.Equal(t, http.StatusNotFound, w.Code)
			},
		},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			step.run(t)
		})
	}
}
