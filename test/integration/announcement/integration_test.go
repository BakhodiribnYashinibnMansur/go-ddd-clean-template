package announcement

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	announcementController "gct/internal/controller/restapi/v1/announcement"
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

func boolPtr(b bool) *bool { return &b }

// ---------------------------------------------------------------------------
// Create
// ---------------------------------------------------------------------------

func TestCreate_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := announcementController.New(useCases.Announcement, setup.TestCfg, l)

	tests := []struct {
		name         string
		body         map[string]any
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name: "success",
			body: map[string]any{
				"title":     "Test Announcement",
				"content":   "This is a test announcement",
				"type":      "info",
				"is_active": true,
			},
			expectedCode: http.StatusCreated,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, "Test Announcement", data["title"])
				assert.Equal(t, "info", data["type"])
				assert.NotEmpty(t, data["id"])
			},
		},
		{
			name: "success - warning type",
			body: map[string]any{
				"title":     "Warning Announcement",
				"content":   "Something is happening",
				"type":      "warning",
				"is_active": false,
			},
			expectedCode: http.StatusCreated,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, "warning", data["type"])
				assert.Equal(t, false, data["is_active"])
			},
		},
		{
			name: "bad request - missing title",
			body: map[string]any{
				"content": "No title",
				"type":    "info",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "bad request - invalid type",
			body: map[string]any{
				"title":   "Bad Type",
				"content": "Invalid type value",
				"type":    "invalid_type",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "bad request - missing content",
			body: map[string]any{
				"title": "No content",
				"type":  "info",
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
	controller := announcementController.New(useCases.Announcement, setup.TestCfg, l)
	ctx := t.Context()

	// Seed an announcement via usecase
	created, err := useCases.Announcement.Create(ctx, createTestAnnouncement("Get Test", "info"))
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
				assert.Equal(t, "Get Test", data["title"])
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
	controller := announcementController.New(useCases.Announcement, setup.TestCfg, l)
	ctx := t.Context()

	// Seed 3 announcements
	for i := range 3 {
		types := []string{"info", "warning", "error"}
		_, err := useCases.Announcement.Create(ctx, createTestAnnouncement("List Announcement "+string(rune('A'+i)), types[i]))
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

	t.Run("filter by type", func(t *testing.T) {
		w, c := newGinContextWithQuery(t, http.MethodGet, "limit=10&offset=0&type=warning", nil)
		controller.List(c)

		assert.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		data, ok := resp["data"].([]any)
		assert.True(t, ok)
		for _, item := range data {
			m := item.(map[string]any)
			assert.Equal(t, "warning", m["type"])
		}
	})

	t.Run("pagination - offset", func(t *testing.T) {
		w, c := newGinContextWithQuery(t, http.MethodGet, "limit=1&offset=0", nil)
		controller.List(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		data := resp["data"].([]any)
		assert.Len(t, data, 1)
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
	controller := announcementController.New(useCases.Announcement, setup.TestCfg, l)
	ctx := t.Context()

	created, err := useCases.Announcement.Create(ctx, createTestAnnouncement("Update Test", "info"))
	require.NoError(t, err)

	tests := []struct {
		name         string
		id           string
		body         map[string]any
		expectedCode int
		checkResp    func(t *testing.T, body []byte)
	}{
		{
			name:         "update title",
			id:           created.ID.String(),
			body:         map[string]any{"title": "Updated Title"},
			expectedCode: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, "Updated Title", data["title"])
			},
		},
		{
			name:         "update type",
			id:           created.ID.String(),
			body:         map[string]any{"type": "error"},
			expectedCode: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				t.Helper()
				var resp map[string]any
				require.NoError(t, json.Unmarshal(body, &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, "error", data["type"])
			},
		},
		{
			name:         "not found",
			id:           uuid.New().String(),
			body:         map[string]any{"title": "Ghost"},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:         "invalid uuid",
			id:           "not-a-uuid",
			body:         map[string]any{"title": "Bad"},
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
	controller := announcementController.New(useCases.Announcement, setup.TestCfg, l)
	ctx := t.Context()

	created, err := useCases.Announcement.Create(ctx, createTestAnnouncement("Delete Test", "info"))
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
				_, err := useCases.Announcement.GetByID(ctx, created.ID)
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
// Toggle
// ---------------------------------------------------------------------------

func TestToggle_Direct(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := announcementController.New(useCases.Announcement, setup.TestCfg, l)
	ctx := t.Context()

	created, err := useCases.Announcement.Create(ctx, createTestAnnouncement("Toggle Test", "info"))
	require.NoError(t, err)
	initialActive := created.IsActive

	t.Run("toggle once", func(t *testing.T) {
		w, c := newGinContext(t, http.MethodPatch, nil)
		c.Params = gin.Params{{Key: "id", Value: created.ID.String()}}

		controller.Toggle(c)

		require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		data := resp["data"].(map[string]any)
		assert.Equal(t, !initialActive, data["is_active"])
	})

	t.Run("toggle back", func(t *testing.T) {
		w, c := newGinContext(t, http.MethodPatch, nil)
		c.Params = gin.Params{{Key: "id", Value: created.ID.String()}}

		controller.Toggle(c)

		require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		data := resp["data"].(map[string]any)
		assert.Equal(t, initialActive, data["is_active"])
	})

	t.Run("invalid uuid", func(t *testing.T) {
		w, c := newGinContext(t, http.MethodPatch, nil)
		c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

		controller.Toggle(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// ---------------------------------------------------------------------------
// Comprehensive Flow
// ---------------------------------------------------------------------------

func TestAnnouncement_ComprehensiveFlow(t *testing.T) {
	cleanDB(t)

	l := logger.New("debug")
	repositories := repo.New(setup.TestPG, setup.TestMinio, setup.TestRedis, &setup.TestCfg.Minio, l)
	useCases := usecase.NewUseCase(repositories, l, setup.TestCfg, nil)
	controller := announcementController.New(useCases.Announcement, setup.TestCfg, l)

	var announcementID string

	steps := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "Step 1: Create",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodPost, map[string]any{
					"title":     "Flow Announcement",
					"content":   "Created in flow test",
					"type":      "success",
					"is_active": true,
				})
				controller.Create(c)
				require.Equal(t, http.StatusCreated, w.Code, "body: %s", w.Body.String())

				var resp map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				data := resp["data"].(map[string]any)
				announcementID = data["id"].(string)
				assert.NotEmpty(t, announcementID)
			},
		},
		{
			name: "Step 2: Get",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodGet, nil)
				c.Params = gin.Params{{Key: "id", Value: announcementID}}

				controller.Get(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				var resp map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, "Flow Announcement", data["title"])
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
				w, c := newGinContext(t, http.MethodPatch, map[string]any{"title": "Updated Flow Announcement"})
				c.Params = gin.Params{{Key: "id", Value: announcementID}}

				controller.Update(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				var resp map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, "Updated Flow Announcement", data["title"])
			},
		},
		{
			name: "Step 5: Toggle",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodPatch, nil)
				c.Params = gin.Params{{Key: "id", Value: announcementID}}

				controller.Toggle(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())

				var resp map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				data := resp["data"].(map[string]any)
				assert.Equal(t, false, data["is_active"])
			},
		},
		{
			name: "Step 6: Delete",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodDelete, nil)
				c.Params = gin.Params{{Key: "id", Value: announcementID}}

				controller.Delete(c)
				require.Equal(t, http.StatusOK, w.Code, "body: %s", w.Body.String())
			},
		},
		{
			name: "Step 7: Verify Deleted",
			run: func(t *testing.T) {
				w, c := newGinContext(t, http.MethodGet, nil)
				c.Params = gin.Params{{Key: "id", Value: announcementID}}

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

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func createTestAnnouncement(title, typ string) domain.CreateAnnouncementRequest {
	return domain.CreateAnnouncementRequest{
		Title:    title,
		Content:  "Test content for " + title,
		Type:     typ,
		IsActive: true,
	}
}
