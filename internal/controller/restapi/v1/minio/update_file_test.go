package minio_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	minioCtrl "gct/internal/controller/restapi/v1/minio"
	"gct/internal/domain"
	"gct/internal/usecase"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateFile(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		id         string
		body       string
		mockSetup  func(*MockFileUseCase)
		wantStatus int
	}{
		{
			name: "success",
			id:   "file-123",
			body: `{"original_name":"updated.png"}`,
			mockSetup: func(m *MockFileUseCase) {
				m.On("UpdateFile", mock.Anything, "file-123", mock.AnythingOfType("domain.UpdateFileMetadataRequest")).
					Return(&domain.FileMetadata{ID: "file-123", OriginalName: "updated.png"}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "bad_request_empty_id",
			id:         "",
			body:       `{"original_name":"updated.png"}`,
			mockSetup:  func(m *MockFileUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad_request_invalid_json",
			id:         "file-123",
			body:       `{invalid}`,
			mockSetup:  func(m *MockFileUseCase) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "internal_error",
			id:   "file-123",
			body: `{"original_name":"updated.png"}`,
			mockSetup: func(m *MockFileUseCase) {
				m.On("UpdateFile", mock.Anything, "file-123", mock.AnythingOfType("domain.UpdateFileMetadataRequest")).
					Return(nil, errors.New("db error"))
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
				{Key: "id", Value: tt.id},
			}

			mockFile := new(MockFileUseCase)
			tt.mockSetup(mockFile)

			uc := &usecase.UseCase{File: mockFile}
			ctrl := minioCtrl.New(uc, log)
			ctrl.UpdateFile(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockFile.AssertExpectations(t)
		})
	}
}
