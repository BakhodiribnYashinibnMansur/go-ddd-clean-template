package http

import (
	"net/http"

	"gct/internal/context/admin/supporting/errorcode"
	"gct/internal/context/admin/supporting/errorcode/application/command"
	"gct/internal/context/admin/supporting/errorcode/application/query"
	"gct/internal/context/admin/supporting/errorcode/domain"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
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

// @Summary Create an error code
// @Description Create a new error code definition
// @Tags ErrorCodes
// @Accept json
// @Produce json
// @Param request body CreateRequest true "Error code data"
// @Success 201 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /error-codes [post]
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

// @Summary List error codes
// @Description Get a paginated list of error codes
// @Tags ErrorCodes
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /error-codes [get]
// List returns a paginated list of error codes.
func (h *Handler) List(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	q := query.ListErrorCodesQuery{
		Filter: domain.ErrorCodeFilter{Limit: pg.Limit, Offset: pg.Offset},
	}
	result, err := h.bc.ListErrorCodes.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.ErrorCodes, "total": result.Total})
}

// @Summary Get an error code
// @Description Get a single error code by ID
// @Tags ErrorCodes
// @Accept json
// @Produce json
// @Param id path string true "Error Code ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /error-codes/{id} [get]
// Get returns a single error code by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := domain.ParseErrorCodeID(ctx.Param("id"))
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

// @Summary Update an error code
// @Description Update an existing error code by ID
// @Tags ErrorCodes
// @Accept json
// @Produce json
// @Param id path string true "Error Code ID"
// @Param request body UpdateRequest true "Error code update data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /error-codes/{id} [patch]
// Update updates an error code.
func (h *Handler) Update(ctx *gin.Context) {
	id, err := domain.ParseErrorCodeID(ctx.Param("id"))
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

// @Summary Delete an error code
// @Description Delete an error code by ID
// @Tags ErrorCodes
// @Accept json
// @Produce json
// @Param id path string true "Error Code ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /error-codes/{id} [delete]
// Delete deletes an error code.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := domain.ParseErrorCodeID(ctx.Param("id"))
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
