package translation_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/translation"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpsert(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		entityType string
		entityID   string
		body       string
		mockSetup  func(*MockUseCase)
		wantStatus int
	}{
		{
			name:       "success",
			entityType: "role",
			entityID:   "550e8400-e29b-41d4-a716-446655440000",
			body:       `{"uz":{"title":"Sarlavha"},"en":{"title":"Title"}}`,
			mockSetup: func(m *MockUseCase) {
				m.On("Upsert", mock.Anything, "role", mock.AnythingOfType("uuid.UUID"), mock.Anything).
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "bad_request_invalid_entity_id",
			entityType: "role",
			entityID:   "not-a-uuid",
			body:       `{"uz":{"title":"Sarlavha"}}`,
			mockSetup:  func(m *MockUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad_request_invalid_json",
			entityType: "role",
			entityID:   "550e8400-e29b-41d4-a716-446655440000",
			body:       `{invalid}`,
			mockSetup:  func(m *MockUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "internal_error",
			entityType: "role",
			entityID:   "550e8400-e29b-41d4-a716-446655440000",
			body:       `{"uz":{"title":"Sarlavha"}}`,
			mockSetup: func(m *MockUseCase) {
				m.On("Upsert", mock.Anything, "role", mock.AnythingOfType("uuid.UUID"), mock.Anything).
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPut, "/", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{
				{Key: "entity_type", Value: tt.entityType},
				{Key: "entity_id", Value: tt.entityID},
			}

			mockUC := new(MockUseCase)
			tt.mockSetup(mockUC)

			ctrl := translation.New(mockUC, cfg, log)
			ctrl.Upsert(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
