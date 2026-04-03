package http

import (
	"net/http"
	"strconv"

	"gct/internal/session"
	"gct/internal/session/application/command"
	appdto "gct/internal/session/application"
	"gct/internal/session/application/query"
	"gct/internal/shared/domain/consts"
	"gct/internal/shared/infrastructure/httpx"
	"gct/internal/shared/infrastructure/httpx/response"
	"gct/internal/shared/infrastructure/logger"

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

// List handles GET /sessions.
func (h *Handler) List(ctx *gin.Context) {
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)
	offset, _ := strconv.ParseInt(ctx.DefaultQuery("offset", "0"), 10, 64)

	filter := appdto.SessionsFilter{
		Limit:  limit,
		Offset: offset,
	}

	if userIDStr := ctx.Query("user_id"); userIDStr != "" {
		uid, err := uuid.Parse(userIDStr)
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

// Get handles GET /sessions/:id.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
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
