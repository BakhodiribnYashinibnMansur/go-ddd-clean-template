package translation_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/translation"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		entityType string
		entityID   string
		lang       string
		mockSetup  func(*MockUseCase)
		wantStatus int
	}{
		{
			name:       "success_delete_all",
			entityType: "role",
			entityID:   "550e8400-e29b-41d4-a716-446655440000",
			lang:       "",
			mockSetup: func(m *MockUseCase) {
				m.On("Delete", mock.Anything, mock.AnythingOfType("domain.TranslationFilter")).
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "success_delete_single_lang",
			entityType: "role",
			entityID:   "550e8400-e29b-41d4-a716-446655440000",
			lang:       "uz",
			mockSetup: func(m *MockUseCase) {
				m.On("Delete", mock.Anything, mock.AnythingOfType("domain.TranslationFilter")).
					Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "bad_request_invalid_entity_id",
			entityType: "role",
			entityID:   "not-a-uuid",
			lang:       "",
			mockSetup:  func(m *MockUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "internal_error",
			entityType: "role",
			entityID:   "550e8400-e29b-41d4-a716-446655440000",
			lang:       "",
			mockSetup: func(m *MockUseCase) {
				m.On("Delete", mock.Anything, mock.AnythingOfType("domain.TranslationFilter")).
					Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			url := "/"
			if tt.lang != "" {
				url += "?lang=" + tt.lang
			}
			c.Request = httptest.NewRequest(http.MethodDelete, url, nil)
			c.Params = gin.Params{
				{Key: "entity_type", Value: tt.entityType},
				{Key: "entity_id", Value: tt.entityID},
			}

			mockUC := new(MockUseCase)
			tt.mockSetup(mockUC)

			ctrl := translation.New(mockUC, cfg, log)
			ctrl.Delete(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUC.AssertExpectations(t)
		})
	}
}
