package util

import (
	"net/http"
	"strconv"
	"time"

	"gct/consts"
	"gct/internal/controller/restapi/response"
	apperrors "gct/pkg/errors"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/gin-gonic/gin"
)

type MockType string

const (
	MockTypeGet    MockType = "get"
	MockTypeGets   MockType = "gets"
	MockTypeCreate MockType = "create"
	MockTypeUpdate MockType = "update"
	MockTypeDelete MockType = "delete"
)

// IsMockMode checks if the current request is in mock mode (triggered via ?mock=true)
func IsMockMode(c *gin.Context) bool {
	mode, exists := c.Get(consts.CtxMockMode)
	if !exists {
		return false
	}
	boolMode, ok := mode.(bool)
	if !ok {
		return false
	}
	return boolMode
}

// MockResponse handles API mocking if mock mode is active.
// It returns true if the mock was handled (response sent).
func MockResponse(c *gin.Context, code int, data, meta any) bool {
	if !IsMockMode(c) {
		return false
	}
	response.ControllerResponse(c, code, data, meta, true)
	return true
}

// GetMocks handles mocking list responses based on pagination and a data generator
func GetMocks(c *gin.Context, generator func(count int) any) bool {
	if !IsMockMode(c) {
		return false
	}
	p, _ := GetPagination(c)
	count := int(p.Limit)
	if count <= 0 {
		count = 10
	}

	data := generator(count)
	total := int64(gofakeit.IntRange(count, count*10))

	meta := &response.Meta{
		Total:  total,
		Limit:  p.Limit,
		Offset: p.Offset,
		Page:   (p.Offset / p.Limit) + 1,
	}

	response.ControllerResponse(c, http.StatusOK, data, meta, true)
	return true
}

// GetMock handles mocking single item responses
func GetMock(c *gin.Context, generator func() any) bool {
	if !IsMockMode(c) {
		return false
	}
	response.ControllerResponse(c, http.StatusOK, generator(), nil, true)
	return true
}

// MockSuccess handles mocking success messages
func MockSuccess(c *gin.Context, message string) bool {
	if !IsMockMode(c) {
		return false
	}
	response.ControllerResponse(c, http.StatusOK, message, nil, true)
	return true
}

// MockCreated handles mocking resource creation
func MockCreated(c *gin.Context, message string) bool {
	if !IsMockMode(c) {
		return false
	}
	response.ControllerResponse(c, http.StatusCreated, message, nil, true)
	return true
}

func MockUpdate(c *gin.Context, message string) bool {
	if !IsMockMode(c) {
		return false
	}
	response.ControllerResponse(c, http.StatusOK, message, nil, true)
	return true
}

func MockDelete(c *gin.Context, message string) bool {
	if !IsMockMode(c) {
		return false
	}
	response.ControllerResponse(c, http.StatusOK, message, nil, true)
	return true
}

// HandleMockDelay introduces an artificial delay if ?mock_delay=ms is provided
func HandleMockDelay(c *gin.Context) {
	if delayStr := c.Query(consts.QueryMockDelay); delayStr != "" {
		if delayMs, err := strconv.Atoi(delayStr); err == nil && delayMs > 0 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}
}

// HandleMockError triggers a specific application error if ?mock_error=CODE is provided
func HandleMockError(c *gin.Context) bool {
	if errCode := c.Query(consts.QueryMockError); errCode != "" {
		appErr := apperrors.New(c.Request.Context(), errCode, "Mock error triggered via query parameter")
		response.RespondWithError(c, appErr, 0)
		c.Abort()
		return true
	}
	return false
}

// HandleMockEmpty returns an empty success response if ?mock_empty=true is provided
func HandleMockEmpty(c *gin.Context) bool {
	if c.Query(consts.QueryMockEmpty) == "true" {
		response.ControllerResponse(c, http.StatusOK, nil, nil, true)
		c.Abort()
		return true
	}
	return false
}

// Mock handles all types of mock responses based on the provided type and payload (generator or message).
func Mock(c *gin.Context, mockType MockType, payload any) bool {
	if !IsMockMode(c) {
		return false
	}

	switch mockType {
	case MockTypeGets:
		if generator, ok := payload.(func(int) any); ok {
			return GetMocks(c, generator)
		}
	case MockTypeGet:
		if generator, ok := payload.(func() any); ok {
			return GetMock(c, generator)
		}
	case MockTypeCreate:
		msg, _ := payload.(string)
		return MockCreated(c, msg)
	case MockTypeUpdate:
		msg, _ := payload.(string)
		return MockUpdate(c, msg)
	case MockTypeDelete:
		msg, _ := payload.(string)
		return MockDelete(c, msg)
	}

	return false
}
