package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gct/consts"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name              string
		existingRequestID string
		shouldGenerate    bool
	}{
		{
			name:              "generate_new_request_id",
			existingRequestID: "",
			shouldGenerate:    true,
		},
		{
			name:              "preserve_existing_request_id",
			existingRequestID: uuid.New().String(),
			shouldGenerate:    false,
		},
		{
			name:              "preserve_custom_request_id",
			existingRequestID: "custom-request-id-12345",
			shouldGenerate:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup router
			r := gin.New()
			r.Use(RequestID())

			var capturedRequestID string
			r.GET("/test", func(c *gin.Context) {
				capturedRequestID = c.GetString(consts.CtxKeyRequestID)
				c.String(http.StatusOK, "ok")
			})

			// Create request
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)

			if tt.existingRequestID != "" {
				req.Header.Set(consts.HeaderXRequestID, tt.existingRequestID)
			}

			// Execute request
			r.ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, http.StatusOK, w.Code)
			assert.NotEmpty(t, capturedRequestID, "Request ID should be set in context")

			if tt.shouldGenerate {
				// Verify it's a valid UUID
				_, err := uuid.Parse(capturedRequestID)
				assert.NoError(t, err, "Generated request ID should be a valid UUID")
			} else {
				// Verify it matches the existing one
				assert.Equal(t, tt.existingRequestID, capturedRequestID, "Should preserve existing request ID")
			}

			// Verify response header is set
			responseRequestID := w.Header().Get(consts.HeaderXRequestID)
			assert.Equal(t, capturedRequestID, responseRequestID, "Response header should match context value")
		})
	}
}
