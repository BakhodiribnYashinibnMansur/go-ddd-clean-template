package http

import (
	"net/http"
	"strconv"

	"gct/internal/context/ops/iprule"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/context/ops/iprule/application/command"
	"gct/internal/context/ops/iprule/application/query"
	"gct/internal/context/ops/iprule/domain"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler provides HTTP endpoints for the IPRule bounded context.
type Handler struct {
	bc *iprule.BoundedContext
	l  logger.Log
}

// NewHandler creates a new IPRule HTTP handler.
func NewHandler(bc *iprule.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// Create creates a new IP rule.
func (h *Handler) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}
	cmd := command.CreateIPRuleCommand{
		IPAddress: req.IPAddress,
		Action:    req.Action,
		Reason:    req.Reason,
		ExpiresAt: req.ExpiresAt,
	}
	if err := h.bc.CreateIPRule.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// List returns a paginated list of IP rules.
func (h *Handler) List(ctx *gin.Context) {
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)
	offset, _ := strconv.ParseInt(ctx.DefaultQuery("offset", "0"), 10, 64)

	q := query.ListIPRulesQuery{
		Filter: domain.IPRuleFilter{Limit: limit, Offset: offset},
	}
	result, err := h.bc.ListIPRules.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.IPRules, "total": result.Total})
}

// Get returns a single IP rule by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	result, err := h.bc.GetIPRule.Handle(ctx.Request.Context(), query.GetIPRuleQuery{ID: domain.IPRuleID(id)})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// Update updates an IP rule.
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
	cmd := command.UpdateIPRuleCommand{
		ID:        domain.IPRuleID(id),
		IPAddress: req.IPAddress,
		Action:    req.Action,
		Reason:    req.Reason,
		ExpiresAt: req.ExpiresAt,
	}
	if err := h.bc.UpdateIPRule.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// Delete deletes an IP rule.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	if err := h.bc.DeleteIPRule.Handle(ctx.Request.Context(), command.DeleteIPRuleCommand{ID: domain.IPRuleID(id)}); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

