package http

import (
	"net/http"
	"strconv"

	"gct/internal/session"
	appdto "gct/internal/session/application"
	"gct/internal/session/application/query"
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
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
			return
		}
		filter.UserID = &uid
	}

	result, err := h.bc.ListSessions.Handle(ctx.Request.Context(), query.ListSessionsQuery{
		Filter: filter,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	view, err := h.bc.GetSession.Handle(ctx.Request.Context(), query.GetSessionQuery{ID: id})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": view})
}
