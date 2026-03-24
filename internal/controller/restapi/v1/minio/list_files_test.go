package minio_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	minioCtrl "gct/internal/controller/restapi/v1/minio"
	"gct/internal/domain"
	"gct/internal/usecase"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListFiles(t *testing.T) {
	gin.SetMode(gin.TestMode)
	log := logger.New(logger.LevelInfo)

	tests := []struct {
		name       string
		query      string
		mockSetup  func(*MockFileUseCase)
		wantStatus int
	}{
		{
			name:  "success",
			query: "?limit=20&offset=0",
			mockSetup: func(m *MockFileUseCase) {
				m.On("ListFiles", mock.Anything, mock.AnythingOfType("domain.FileMetadataFilter")).
					Return([]domain.FileMetadata{{ID: "f1", OriginalName: "test.png"}}, int64(1), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "success_with_filters",
			query: "?limit=10&offset=0&search=test&mime_type=image/png",
			mockSetup: func(m *MockFileUseCase) {
				m.On("ListFiles", mock.Anything, mock.AnythingOfType("domain.FileMetadataFilter")).
					Return([]domain.FileMetadata{{ID: "f1", OriginalName: "test.png"}}, int64(1), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "success_default_pagination",
			query: "",
			mockSetup: func(m *MockFileUseCase) {
				m.On("ListFiles", mock.Anything, mock.AnythingOfType("domain.FileMetadataFilter")).
					Return([]domain.FileMetadata{}, int64(0), nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "internal_error",
			query: "?limit=20&offset=0",
			mockSetup: func(m *MockFileUseCase) {
				m.On("ListFiles", mock.Anything, mock.AnythingOfType("domain.FileMetadataFilter")).
					Return(nil, int64(0), errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/"+tt.query, nil)

			mockFile := new(MockFileUseCase)
			tt.mockSetup(mockFile)

			uc := &usecase.UseCase{File: mockFile}
			ctrl := minioCtrl.New(uc, log)
			ctrl.ListFiles(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockFile.AssertExpectations(t)
		})
	}
}
