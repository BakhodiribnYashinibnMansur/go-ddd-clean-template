package http

import (
	"net/http"
	"strconv"

	"gct/internal/context/content/notification"
	"gct/internal/platform/infrastructure/httpx"
	"gct/internal/platform/infrastructure/httpx/response"
	"gct/internal/context/content/notification/application/command"
	"gct/internal/context/content/notification/application/query"
	"gct/internal/context/content/notification/domain"
	"gct/internal/platform/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler provides HTTP endpoints for the Notification bounded context.
type Handler struct {
	bc *notification.BoundedContext
	l  logger.Log
}

// NewHandler creates a new Notification HTTP handler.
func NewHandler(bc *notification.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// Create creates a new notification.
func (h *Handler) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.CreateCommand{
		UserID:  req.UserID,
		Title:   req.Title,
		Message: req.Message,
		Type:    req.Type,
	}
	if err := h.bc.CreateNotification.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// List returns a paginated list of notifications.
func (h *Handler) List(ctx *gin.Context) {
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)
	offset, _ := strconv.ParseInt(ctx.DefaultQuery("offset", "0"), 10, 64)

	q := query.ListQuery{
		Filter: domain.NotificationFilter{Limit: limit, Offset: offset},
	}
	result, err := h.bc.ListNotifications.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Notifications, "total": result.Total})
}

// Get returns a single notification by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	result, err := h.bc.GetNotification.Handle(ctx.Request.Context(), query.GetQuery{ID: id})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// Delete deletes a notification.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	if err := h.bc.DeleteNotification.Handle(ctx.Request.Context(), command.DeleteCommand{ID: id}); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
