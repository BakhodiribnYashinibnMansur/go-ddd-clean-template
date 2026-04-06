package http

import (
	"net/http"

	"gct/internal/context/iam/generic/authz"
	"gct/internal/context/iam/generic/authz/application/command"
	"gct/internal/context/iam/generic/authz/application/query"
	"gct/internal/context/iam/generic/authz/domain"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler holds dependencies for Authz HTTP handlers.
type Handler struct {
	bc *authz.BoundedContext
	l  logger.Log
}

// NewHandler creates a new Authz HTTP handler.
func NewHandler(bc *authz.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// --- Roles ---

// @Summary Create a role
// @Description Create a new role
// @Tags Roles
// @Accept json
// @Produce json
// @Param request body CreateRoleRequest true "Role data"
// @Success 201 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /roles [post]
// CreateRole handles POST /roles.
func (h *Handler) CreateRole(ctx *gin.Context) {
	var req CreateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	err := h.bc.CreateRole.Handle(ctx.Request.Context(), command.CreateRoleCommand{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// @Summary List roles
// @Description Get a paginated list of roles
// @Tags Roles
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /roles [get]
// ListRoles handles GET /roles.
func (h *Handler) ListRoles(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	result, err := h.bc.ListRoles.Handle(ctx.Request.Context(), query.ListRolesQuery{
		Pagination: pg,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  result.Roles,
		"total": result.Total,
	})
}

// @Summary Get a role
// @Description Get role details by ID
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /roles/{id} [get]
// GetRole handles GET /roles/:id.
func (h *Handler) GetRole(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	view, err := h.bc.GetRole.Handle(ctx.Request.Context(), query.GetRoleQuery{ID: domain.RoleID(id)})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": view})
}

// @Summary Update a role
// @Description Update role details by ID
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID (UUID)"
// @Param request body UpdateRoleRequest true "Role update data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /roles/{id} [patch]
// UpdateRole handles PATCH /roles/:id.
func (h *Handler) UpdateRole(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	var req UpdateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	err = h.bc.UpdateRole.Handle(ctx.Request.Context(), command.UpdateRoleCommand{
		ID:          domain.RoleID(id),
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// @Summary Delete a role
// @Description Delete a role by ID
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID (UUID)"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /roles/{id} [delete]
// DeleteRole handles DELETE /roles/:id.
func (h *Handler) DeleteRole(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	err = h.bc.DeleteRole.Handle(ctx.Request.Context(), command.DeleteRoleCommand{ID: domain.RoleID(id)})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// --- Permissions ---

// @Summary Create a permission
// @Description Create a new permission
// @Tags Permissions
// @Accept json
// @Produce json
// @Param request body CreatePermissionRequest true "Permission data"
// @Success 201 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /permissions [post]
// CreatePermission handles POST /permissions.
func (h *Handler) CreatePermission(ctx *gin.Context) {
	var req CreatePermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	var parentID *domain.PermissionID
	if req.ParentID != nil {
		pid := domain.PermissionID(*req.ParentID)
		parentID = &pid
	}
	err := h.bc.CreatePermission.Handle(ctx.Request.Context(), command.CreatePermissionCommand{
		Name:        req.Name,
		ParentID:    parentID,
		Description: req.Description,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// @Summary List permissions
// @Description Get a paginated list of permissions
// @Tags Permissions
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /permissions [get]
// ListPermissions handles GET /permissions.
func (h *Handler) ListPermissions(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	result, err := h.bc.ListPermissions.Handle(ctx.Request.Context(), query.ListPermissionsQuery{
		Pagination: pg,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  result.Permissions,
		"total": result.Total,
	})
}

// @Summary Delete a permission
// @Description Delete a permission by ID
// @Tags Permissions
// @Accept json
// @Produce json
// @Param id path string true "Permission ID (UUID)"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /permissions/{id} [delete]
// DeletePermission handles DELETE /permissions/:id.
func (h *Handler) DeletePermission(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	err = h.bc.DeletePermission.Handle(ctx.Request.Context(), command.DeletePermissionCommand{ID: domain.PermissionID(id)})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// --- Policies ---

// @Summary Create a policy
// @Description Create a new policy
// @Tags Policies
// @Accept json
// @Produce json
// @Param request body CreatePolicyRequest true "Policy data"
// @Success 201 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /policies [post]
// CreatePolicy handles POST /policies.
func (h *Handler) CreatePolicy(ctx *gin.Context) {
	var req CreatePolicyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	err := h.bc.CreatePolicy.Handle(ctx.Request.Context(), command.CreatePolicyCommand{
		PermissionID: req.PermissionID,
		Effect:       req.Effect,
		Priority:     req.Priority,
		Conditions:   req.Conditions,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// @Summary List policies
// @Description Get a paginated list of policies
// @Tags Policies
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /policies [get]
// ListPolicies handles GET /policies.
func (h *Handler) ListPolicies(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	result, err := h.bc.ListPolicies.Handle(ctx.Request.Context(), query.ListPoliciesQuery{
		Pagination: pg,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  result.Policies,
		"total": result.Total,
	})
}

// @Summary Update a policy
// @Description Update policy details by ID
// @Tags Policies
// @Accept json
// @Produce json
// @Param id path string true "Policy ID (UUID)"
// @Param request body UpdatePolicyRequest true "Policy update data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /policies/{id} [patch]
// UpdatePolicy handles PATCH /policies/:id.
func (h *Handler) UpdatePolicy(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	var req UpdatePolicyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	err = h.bc.UpdatePolicy.Handle(ctx.Request.Context(), command.UpdatePolicyCommand{
		ID:         domain.PolicyID(id),
		Effect:     req.Effect,
		Priority:   req.Priority,
		Conditions: req.Conditions,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// @Summary Delete a policy
// @Description Delete a policy by ID
// @Tags Policies
// @Accept json
// @Produce json
// @Param id path string true "Policy ID (UUID)"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /policies/{id} [delete]
// DeletePolicy handles DELETE /policies/:id.
func (h *Handler) DeletePolicy(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	err = h.bc.DeletePolicy.Handle(ctx.Request.Context(), command.DeletePolicyCommand{ID: domain.PolicyID(id)})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// @Summary Toggle a policy
// @Description Toggle the enabled state of a policy by ID
// @Tags Policies
// @Accept json
// @Produce json
// @Param id path string true "Policy ID (UUID)"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /policies/{id}/toggle [post]
// TogglePolicy handles POST /policies/:id/toggle.
func (h *Handler) TogglePolicy(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	err = h.bc.TogglePolicy.Handle(ctx.Request.Context(), command.TogglePolicyCommand{ID: domain.PolicyID(id)})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// --- Scopes ---

// @Summary Create a scope
// @Description Create a new scope
// @Tags Scopes
// @Accept json
// @Produce json
// @Param request body CreateScopeRequest true "Scope data"
// @Success 201 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /scopes [post]
// CreateScope handles POST /scopes.
func (h *Handler) CreateScope(ctx *gin.Context) {
	var req CreateScopeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	err := h.bc.CreateScope.Handle(ctx.Request.Context(), command.CreateScopeCommand{
		Path:   req.Path,
		Method: req.Method,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// @Summary List scopes
// @Description Get a paginated list of scopes
// @Tags Scopes
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /scopes [get]
// ListScopes handles GET /scopes.
func (h *Handler) ListScopes(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	result, err := h.bc.ListScopes.Handle(ctx.Request.Context(), query.ListScopesQuery{
		Pagination: pg,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  result.Scopes,
		"total": result.Total,
	})
}

// @Summary Delete a scope
// @Description Delete a scope by path and method
// @Tags Scopes
// @Accept json
// @Produce json
// @Param request body DeleteScopeRequest true "Scope identification data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /scopes [delete]
// DeleteScope handles DELETE /scopes.
func (h *Handler) DeleteScope(ctx *gin.Context) {
	var req DeleteScopeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	err := h.bc.DeleteScope.Handle(ctx.Request.Context(), command.DeleteScopeCommand{
		Path:   req.Path,
		Method: req.Method,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// --- Assignments ---

// @Summary Assign permission to role
// @Description Assign a permission to a role by role ID
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID (UUID)"
// @Param request body AssignPermissionRequest true "Permission assignment data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /roles/{id}/permissions [post]
// AssignPermission handles POST /roles/:id/permissions.
func (h *Handler) AssignPermission(ctx *gin.Context) {
	roleID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	var req AssignPermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	err = h.bc.AssignPermission.Handle(ctx.Request.Context(), command.AssignPermissionCommand{
		RoleID:       domain.RoleID(roleID),
		PermissionID: req.PermissionID,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// @Summary Assign scope to permission
// @Description Assign a scope to a permission by permission ID
// @Tags Permissions
// @Accept json
// @Produce json
// @Param id path string true "Permission ID (UUID)"
// @Param request body AssignScopeRequest true "Scope assignment data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /permissions/{id}/scopes [post]
// AssignScope handles POST /permissions/:id/scopes.
func (h *Handler) AssignScope(ctx *gin.Context) {
	permID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	var req AssignScopeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	err = h.bc.AssignScope.Handle(ctx.Request.Context(), command.AssignScopeCommand{
		PermissionID: domain.PermissionID(permID),
		Path:         req.Path,
		Method:       req.Method,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
