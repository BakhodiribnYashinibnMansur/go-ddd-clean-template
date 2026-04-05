package http

import (
	"net/http"
	"strconv"

	"gct/internal/context/admin/errorcode"
	"gct/internal/context/admin/errorcode/application/command"
	"gct/internal/context/admin/errorcode/application/query"
	"gct/internal/context/admin/errorcode/domain"
	"gct/internal/platform/infrastructure/httpx"
	"gct/internal/platform/infrastructure/httpx/response"
	"gct/internal/platform/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler provides HTTP endpoints for the ErrorCode bounded context.
type Handler struct {
	bc *errorcode.BoundedContext
	l  logger.Log
}

// NewHandler creates a new ErrorCode HTTP handler.
func NewHandler(bc *errorcode.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// Create creates a new error code.
func (h *Handler) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.CreateErrorCodeCommand{
		Code:       req.Code,
		Message:    req.Message,
		MessageUz:  req.MessageUz,
		MessageRu:  req.MessageRu,
		HTTPStatus: req.HTTPStatus,
		Category:   req.Category,
		Severity:   req.Severity,
		Retryable:  req.Retryable,
		RetryAfter: req.RetryAfter,
		Suggestion: req.Suggestion,
	}
	if err := h.bc.CreateErrorCode.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// List returns a paginated list of error codes.
func (h *Handler) List(ctx *gin.Context) {
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)
	offset, _ := strconv.ParseInt(ctx.DefaultQuery("offset", "0"), 10, 64)

	q := query.ListErrorCodesQuery{
		Filter: domain.ErrorCodeFilter{Limit: limit, Offset: offset},
	}
	result, err := h.bc.ListErrorCodes.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.ErrorCodes, "total": result.Total})
}

// Get returns a single error code by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	result, err := h.bc.GetErrorCode.Handle(ctx.Request.Context(), query.GetErrorCodeQuery{ID: id})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// Update updates an error code.
func (h *Handler) Update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	var req UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.UpdateErrorCodeCommand{
		ID:         id,
		Message:    req.Message,
		MessageUz:  req.MessageUz,
		MessageRu:  req.MessageRu,
		HTTPStatus: req.HTTPStatus,
		Category:   req.Category,
		Severity:   req.Severity,
		Retryable:  req.Retryable,
		RetryAfter: req.RetryAfter,
		Suggestion: req.Suggestion,
	}
	if err := h.bc.UpdateErrorCode.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// Delete deletes an error code.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	if err := h.bc.DeleteErrorCode.Handle(ctx.Request.Context(), command.DeleteErrorCodeCommand{ID: id}); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
