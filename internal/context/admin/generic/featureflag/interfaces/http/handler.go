package http

import (
	"net/http"

	"gct/internal/context/admin/generic/featureflag"
	"gct/internal/context/admin/generic/featureflag/application/command"
	"gct/internal/context/admin/generic/featureflag/application/query"
	"gct/internal/context/admin/generic/featureflag/domain"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"

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

// @Summary Create a feature flag
// @Description Create a new feature flag
// @Tags FeatureFlags
// @Accept json
// @Produce json
// @Param request body CreateRequest true "Feature flag data"
// @Success 201 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /feature-flags [post]
// Create creates a new feature flag.
func (h *Handler) Create(ctx *gin.Context) {
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
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
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// @Summary List feature flags
// @Description Get a paginated list of feature flags
// @Tags FeatureFlags
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /feature-flags [get]
// List returns a paginated list of feature flags.
func (h *Handler) List(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	q := query.ListQuery{
		Filter: domain.FeatureFlagFilter{Limit: pg.Limit, Offset: pg.Offset},
	}
	result, err := h.bc.ListFlags.Handle(ctx.Request.Context(), q)
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result.Flags, "total": result.Total})
}

// @Summary Get a feature flag
// @Description Get a single feature flag by ID
// @Tags FeatureFlags
// @Accept json
// @Produce json
// @Param id path string true "Feature Flag ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /feature-flags/{id} [get]
// Get returns a single feature flag by ID.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := domain.ParseFeatureFlagID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	result, err := h.bc.GetFlag.Handle(ctx.Request.Context(), query.GetQuery{ID: id})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

// @Summary Update a feature flag
// @Description Update an existing feature flag by ID
// @Tags FeatureFlags
// @Accept json
// @Produce json
// @Param id path string true "Feature Flag ID"
// @Param request body UpdateRequest true "Feature flag update data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /feature-flags/{id} [patch]
// Update updates a feature flag.
func (h *Handler) Update(ctx *gin.Context) {
	id, err := domain.ParseFeatureFlagID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	var req UpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
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
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// @Summary Delete a feature flag
// @Description Delete a feature flag by ID
// @Tags FeatureFlags
// @Accept json
// @Produce json
// @Param id path string true "Feature Flag ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /feature-flags/{id} [delete]
// Delete deletes a feature flag.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := domain.ParseFeatureFlagID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	if err := h.bc.DeleteFlag.Handle(ctx.Request.Context(), command.DeleteCommand{ID: id}); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// @Summary Create a rule group
// @Description Add a rule group to a feature flag
// @Tags FeatureFlags
// @Accept json
// @Produce json
// @Param id path string true "Feature Flag ID"
// @Param request body CreateRuleGroupRequest true "Rule group data"
// @Success 201 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /feature-flags/{id}/rule-groups [post]
// CreateRuleGroup adds a rule group to a feature flag.
func (h *Handler) CreateRuleGroup(ctx *gin.Context) {
	flagID, err := domain.ParseFeatureFlagID(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	var req CreateRuleGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
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
		FlagID:     domain.FeatureFlagID(flagID),
		Name:       req.Name,
		Variation:  req.Variation,
		Priority:   req.Priority,
		Conditions: conditions,
	}
	if err := h.bc.CreateRuleGroup.Handle(ctx.Request.Context(), cmd); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// @Summary Update a rule group
// @Description Update an existing rule group of a feature flag
// @Tags FeatureFlags
// @Accept json
// @Produce json
// @Param id path string true "Feature Flag ID"
// @Param groupId path string true "Rule Group ID"
// @Param request body UpdateRuleGroupRequest true "Rule group update data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /feature-flags/{id}/rule-groups/{groupId} [patch]
// UpdateRuleGroup updates an existing rule group.
func (h *Handler) UpdateRuleGroup(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("groupId"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	var req UpdateRuleGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	cmd := command.UpdateRuleGroupCommand{
		ID:        domain.RuleGroupID(groupID),
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
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// @Summary Delete a rule group
// @Description Remove a rule group from a feature flag
// @Tags FeatureFlags
// @Accept json
// @Produce json
// @Param id path string true "Feature Flag ID"
// @Param groupId path string true "Rule Group ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /feature-flags/{id}/rule-groups/{groupId} [delete]
// DeleteRuleGroup removes a rule group.
func (h *Handler) DeleteRuleGroup(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("groupId"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}
	if err := h.bc.DeleteRuleGroup.Handle(ctx.Request.Context(), command.DeleteRuleGroupCommand{ID: domain.RuleGroupID(groupID)}); err != nil {
		response.HandleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// @Summary Evaluate a feature flag
// @Description Evaluate a single feature flag for the given user attributes
// @Tags FeatureFlags
// @Accept json
// @Produce json
// @Param request body EvaluateRequest true "Evaluation request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /feature-flags/evaluate [post]
// Evaluate evaluates a single feature flag for the given user attributes.
func (h *Handler) Evaluate(ctx *gin.Context) {
	var req EvaluateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	result, err := h.bc.EvaluateFlag.Handle(ctx.Request.Context(), query.EvaluateQuery{
		Key:       req.Key,
		UserAttrs: req.UserAttrs,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"key":       result.Key,
		"value":     result.Value,
		"flag_type": result.FlagType,
	})
}

// @Summary Batch evaluate feature flags
// @Description Evaluate multiple feature flags for the given user attributes
// @Tags FeatureFlags
// @Accept json
// @Produce json
// @Param request body BatchEvaluateRequest true "Batch evaluation request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /feature-flags/batch-evaluate [post]
// BatchEvaluate evaluates multiple feature flags for the given user attributes.
func (h *Handler) BatchEvaluate(ctx *gin.Context) {
	var req BatchEvaluateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	result, err := h.bc.BatchEvaluateFlag.Handle(ctx.Request.Context(), query.BatchEvaluateQuery{
		Keys:      req.Keys,
		UserAttrs: req.UserAttrs,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	flags := make(map[string]gin.H, len(result.Flags))
	for k, v := range result.Flags {
		flags[k] = gin.H{"value": v.Value, "flag_type": v.FlagType}
	}

	ctx.JSON(http.StatusOK, gin.H{"flags": flags})
}
