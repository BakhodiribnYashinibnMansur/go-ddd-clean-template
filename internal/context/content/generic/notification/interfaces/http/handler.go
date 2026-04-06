package http

import (
	"net/http"

	"gct/internal/context/content/generic/notification"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/context/content/generic/notification/application/command"
	"gct/internal/context/content/generic/notification/application/query"
	"gct/internal/context/content/generic/notification/domain"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
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

// @Summary Create a notification
// @Description Create a new notification
// @Tags Notifications
// @Accept json
// @Produce json
// @Param request body CreateRequest true "Notification data"
// @Success 201 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /notifications [post]
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

// @Summary List notifications
// @Description Get a paginated list of notifications
// @Tags Notifications
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /notifications [get]
// List returns a paginated list of notifications.
func (h *Handler) List(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	q := query.ListQuery{
		Filter: domain.NotificationFilter{Limit: pg.Limit, Offset: pg.Offset},
	}
	result, err := h.bc.ListNotifications.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Notifications, "total": result.Total})
}

// @Summary Get a notification
// @Description Get a notification by ID
// @Tags Notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /notifications/{id} [get]
// Get returns a single notification by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := domain.ParseNotificationID(ctx.Param("id"))
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

// @Summary Delete a notification
// @Description Delete a notification by ID
// @Tags Notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /notifications/{id} [delete]
// Delete deletes a notification.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := domain.ParseNotificationID(ctx.Param("id"))
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
