package http

import (
	"net/http"

	"gct/internal/context/iam/generic/session"
	appdto "gct/internal/context/iam/generic/session/application"
	"gct/internal/context/iam/generic/session/application/command"
	"gct/internal/context/iam/generic/session/application/query"
	sessiondomain "gct/internal/context/iam/generic/session/domain"
	"gct/internal/kernel/consts"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler holds dependencies for Session HTTP handlers.
type Handler struct {
	bc *session.BoundedContext
	l  logger.Log
}

// NewHandler creates a new Session HTTP handler.
func NewHandler(bc *session.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// @Summary List sessions
// @Description Get a paginated list of sessions with optional user filter
// @Tags Sessions
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Param user_id query string false "Filter by user ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /sessions [get]
// List handles GET /sessions.
func (h *Handler) List(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	filter := appdto.SessionsFilter{
		Limit:  pg.Limit,
		Offset: pg.Offset,
	}

	if userIDStr := ctx.Query("user_id"); userIDStr != "" {
		uid, err := sessiondomain.ParseUserID(userIDStr)
		if err != nil {
			response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
			return
		}
		filter.UserID = &uid
	}

	result, err := h.bc.ListSessions.Handle(ctx.Request.Context(), query.ListSessionsQuery{
		Filter: filter,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  result.Sessions,
		"total": result.Total,
	})
}

// @Summary Get a session
// @Description Get session details by ID
// @Tags Sessions
// @Accept json
// @Produce json
// @Param id path string true "Session ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /sessions/{id} [get]
// Get handles GET /sessions/:id.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := sessiondomain.ParseSessionID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	view, err := h.bc.GetSession.Handle(ctx.Request.Context(), query.GetSessionQuery{ID: id})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": view})
}

// @Summary Delete a session
// @Description Revoke a session by ID
// @Tags Sessions
// @Accept json
// @Produce json
// @Param id path string true "Session ID (UUID)"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /sessions/{id} [delete]
// Delete handles DELETE /sessions/:id.
func (h *Handler) Delete(ctx *gin.Context) {
	sessionID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	userIDStr, exists := ctx.Get(consts.CtxUserID)
	if !exists {
		response.RespondWithError(ctx, httpx.ErrUnAuth, http.StatusUnauthorized)
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrUserIdNotFound, http.StatusUnauthorized)
		return
	}

	if err := h.bc.RevokeSession.Handle(ctx.Request.Context(), command.RevokeSessionCommand{
		UserID:    userID,
		SessionID: sessionID,
	}); err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// @Summary Revoke all sessions
// @Description Revoke all sessions for the authenticated user
// @Tags Sessions
// @Accept json
// @Produce json
// @Success 200 {object} map[string]bool
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /sessions/revoke-all [post]
// RevokeAll handles POST /sessions/revoke-all.
func (h *Handler) RevokeAll(ctx *gin.Context) {
	userIDStr, exists := ctx.Get(consts.CtxUserID)
	if !exists {
		response.RespondWithError(ctx, httpx.ErrUnAuth, http.StatusUnauthorized)
		return
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrUserIdNotFound, http.StatusUnauthorized)
		return
	}

	if err := h.bc.RevokeAllSessions.Handle(ctx.Request.Context(), command.RevokeAllSessionsCommand{
		UserID: userID,
	}); err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
