package minio_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	minioCtrl "gct/internal/controller/restapi/v1/minio"
	"gct/internal/usecase"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		id         string
		mockSetup  func(*MockFileUseCase)
		wantStatus int
	}{
		{
			name: "success",
			id:   "file-123",
			mockSetup: func(m *MockFileUseCase) {
				m.On("DeleteFile", mock.Anything, "file-123").Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "bad_request_empty_id",
			id:         "",
			mockSetup:  func(m *MockFileUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "internal_error",
			id:   "file-123",
			mockSetup: func(m *MockFileUseCase) {
				m.On("DeleteFile", mock.Anything, "file-123").Return(errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/", nil)
			c.Params = gin.Params{
				{Key: "id", Value: tt.id},
			}

			mockFile := new(MockFileUseCase)
			tt.mockSetup(mockFile)

			uc := &usecase.UseCase{File: mockFile}
			ctrl := minioCtrl.New(uc, log)
			ctrl.DeleteFile(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockFile.AssertExpectations(t)
		})
	}
}
