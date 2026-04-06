package http

import (
	"net/http"

	"gct/internal/context/ops/supporting/iprule"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/context/ops/supporting/iprule/application/command"
	"gct/internal/context/ops/supporting/iprule/application/query"
	ipruleentity "gct/internal/context/ops/supporting/iprule/domain/entity"
	iprulerepo "gct/internal/context/ops/supporting/iprule/domain/repository"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
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

// @Summary Create an IP rule
// @Description Create a new IP rule
// @Tags IPRules
// @Accept json
// @Produce json
// @Param request body CreateRequest true "IP rule data"
// @Success 201 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /ip-rules [post]
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

// @Summary List IP rules
// @Description Return a paginated list of IP rules
// @Tags IPRules
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /ip-rules [get]
// List returns a paginated list of IP rules.
func (h *Handler) List(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	q := query.ListIPRulesQuery{
		Filter: iprulerepo.IPRuleFilter{Limit: pg.Limit, Offset: pg.Offset},
	}
	result, err := h.bc.ListIPRules.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.IPRules, "total": result.Total})
}

// @Summary Get an IP rule
// @Description Return a single IP rule by ID
// @Tags IPRules
// @Accept json
// @Produce json
// @Param id path string true "IP rule ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /ip-rules/{id} [get]
// Get returns a single IP rule by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := ipruleentity.ParseIPRuleID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	result, err := h.bc.GetIPRule.Handle(ctx.Request.Context(), query.GetIPRuleQuery{ID: id})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// @Summary Update an IP rule
// @Description Update an existing IP rule
// @Tags IPRules
// @Accept json
// @Produce json
// @Param id path string true "IP rule ID"
// @Param request body UpdateRequest true "IP rule update data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /ip-rules/{id} [patch]
// Update updates an IP rule.
func (h *Handler) Update(ctx *gin.Context) {
	id, err := ipruleentity.ParseIPRuleID(ctx.Param("id"))
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
		ID:        id,
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

// @Summary Delete an IP rule
// @Description Delete an IP rule by ID
// @Tags IPRules
// @Accept json
// @Produce json
// @Param id path string true "IP rule ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /ip-rules/{id} [delete]
// Delete deletes an IP rule.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := ipruleentity.ParseIPRuleID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	if err := h.bc.DeleteIPRule.Handle(ctx.Request.Context(), command.DeleteIPRuleCommand{ID: id}); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

