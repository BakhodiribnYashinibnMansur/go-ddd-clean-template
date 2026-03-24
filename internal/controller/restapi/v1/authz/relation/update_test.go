package relation_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gct/config"
	"gct/internal/controller/restapi/v1/authz/relation"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	log := logger.New(logger.LevelInfo)
	validID := uuid.New()

	tests := []struct {
		name       string
		paramID    string
		body       string
		mockSetup  func(*MockRelationUC)
		wantStatus int
	}{
		{
			name:    "success",
			paramID: validID.String(),
			body:    `{"type":"REGION","name":"West"}`,
			mockSetup: func(m *MockRelationUC) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.Relation")).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "bad_request_invalid_uuid",
			paramID: "bad",
			body:    `{"type":"REGION","name":"West"}`,
			mockSetup: func(m *MockRelationUC) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "bad_request_invalid_json",
			paramID: validID.String(),
			body:    `{invalid`,
			mockSetup: func(m *MockRelationUC) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "internal_error",
			paramID: validID.String(),
			body:    `{"type":"REGION","name":"West"}`,
			mockSetup: func(m *MockRelationUC) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*domain.Relation")).
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
			c.Params = gin.Params{{Key: "relation_id", Value: tt.paramID}}

			mockRelation := new(MockRelationUC)
			tt.mockSetup(mockRelation)

			ctrl := relation.New(newTestUseCase(mockRelation), cfg, log)
			ctrl.Update(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockRelation.AssertExpectations(t)
		})
	}
}
