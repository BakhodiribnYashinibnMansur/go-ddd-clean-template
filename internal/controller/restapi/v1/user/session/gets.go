package session

import (
	"net/http"
	"strconv"

	"github.com/evrone/go-clean-template/consts"
	"github.com/evrone/go-clean-template/internal/controller/restapi/response"
	"github.com/gin-gonic/gin"
)

// GetActiveSessions godoc
// @Summary     Get active sessions
// @Description List all active sessions for the current user
// @Tags        sessions
// @Accept      json
// @Produce     json
// @Param       page  query int false "Page number"
// @Param       limit query int false "Page size"
// @Success     200 {object} response.SuccessResponse
// @Failure     401 {object} response.ErrorResponse
// @Failure     500 {object} response.ErrorResponse
// @Router      /sessions [get]
func (c *Controller) Sessions(ctx *gin.Context) {
	_, exists := ctx.Get(consts.CtxUserID)
	if !exists {
		response.ControllerResponse(ctx, http.StatusUnauthorized, "unauthorized", nil, false)
		return
	}

	page := 1
	limit := 10

	if p := ctx.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := ctx.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Placeholder
	responseData := map[string]any{
		"sessions": []any{},
		"total":    0,
		"page":     page,
		"limit":    limit,
	}

	response.ControllerResponse(ctx, http.StatusOK, responseData, nil, true)
}
