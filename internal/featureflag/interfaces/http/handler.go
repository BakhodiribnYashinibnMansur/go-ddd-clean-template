package http

import (
	"net/http"
	"strconv"

	"gct/internal/featureflag"
	"gct/internal/featureflag/application/command"
	"gct/internal/featureflag/application/query"
	"gct/internal/featureflag/domain"
	"gct/internal/shared/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler provides HTTP endpoints for the FeatureFlag bounded context.
type Handler struct {
	bc *featureflag.BoundedContext
	l  logger.Log
}

// NewHandler creates a new FeatureFlag HTTP handler.
func NewHandler(bc *featureflag.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// Create creates a new feature flag.
func (h *Handler) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cmd := command.CreateCommand{
		Name:              req.Name,
		Key:               req.Key,
		Description:       req.Description,
		FlagType:          req.FlagType,
		DefaultValue:      req.DefaultValue,
		RolloutPercentage: req.RolloutPercentage,
		IsActive:          req.IsActive,
	}
	if err := h.bc.CreateFlag.Handle(ctx.Request.Context(), cmd); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// List returns a paginated list of feature flags.
func (h *Handler) List(ctx *gin.Context) {
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)
	offset, _ := strconv.ParseInt(ctx.DefaultQuery("offset", "0"), 10, 64)

	q := query.ListQuery{
		Filter: domain.FeatureFlagFilter{Limit: limit, Offset: offset},
	}
	result, err := h.bc.ListFlags.Handle(ctx.Request.Context(), q)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Flags, "total": result.Total})
}

// Get returns a single feature flag by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	result, err := h.bc.GetFlag.Handle(ctx.Request.Context(), query.GetQuery{ID: id})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// Update updates a feature flag.
func (h *Handler) Update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cmd := command.UpdateCommand{
		ID:                id,
		Name:              req.Name,
		Key:               req.Key,
		Description:       req.Description,
		FlagType:          req.FlagType,
		DefaultValue:      req.DefaultValue,
		RolloutPercentage: req.RolloutPercentage,
		IsActive:          req.IsActive,
	}
	if err := h.bc.UpdateFlag.Handle(ctx.Request.Context(), cmd); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// Delete deletes a feature flag.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.bc.DeleteFlag.Handle(ctx.Request.Context(), command.DeleteCommand{ID: id}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// CreateRuleGroup adds a rule group to a feature flag.
func (h *Handler) CreateRuleGroup(ctx *gin.Context) {
	flagID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req CreateRuleGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conditions := make([]command.ConditionInput, len(req.Conditions))
	for i, c := range req.Conditions {
		conditions[i] = command.ConditionInput{
			Attribute: c.Attribute,
			Operator:  c.Operator,
			Value:     c.Value,
		}
	}

	cmd := command.CreateRuleGroupCommand{
		FlagID:     flagID,
		Name:       req.Name,
		Variation:  req.Variation,
		Priority:   req.Priority,
		Conditions: conditions,
	}
	if err := h.bc.CreateRuleGroup.Handle(ctx.Request.Context(), cmd); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// UpdateRuleGroup updates an existing rule group.
func (h *Handler) UpdateRuleGroup(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("groupId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid groupId"})
		return
	}
	var req UpdateRuleGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := command.UpdateRuleGroupCommand{
		ID:        groupID,
		Name:      req.Name,
		Variation: req.Variation,
		Priority:  req.Priority,
	}

	if req.Conditions != nil {
		conditions := make([]command.ConditionInput, len(*req.Conditions))
		for i, c := range *req.Conditions {
			conditions[i] = command.ConditionInput{
				Attribute: c.Attribute,
				Operator:  c.Operator,
				Value:     c.Value,
			}
		}
		cmd.Conditions = &conditions
	}

	if err := h.bc.UpdateRuleGroup.Handle(ctx.Request.Context(), cmd); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteRuleGroup removes a rule group.
func (h *Handler) DeleteRuleGroup(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("groupId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid groupId"})
		return
	}
	if err := h.bc.DeleteRuleGroup.Handle(ctx.Request.Context(), command.DeleteRuleGroupCommand{ID: groupID}); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
