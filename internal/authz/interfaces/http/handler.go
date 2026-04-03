package http

import (
	"net/http"

	"gct/internal/authz"
	"gct/internal/authz/application/command"
	"gct/internal/authz/application/query"
	"gct/internal/shared/infrastructure/httpx"
	"gct/internal/shared/infrastructure/httpx/response"
	"gct/internal/shared/infrastructure/logger"

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

// GetRole handles GET /roles/:id.
func (h *Handler) GetRole(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	view, err := h.bc.GetRole.Handle(ctx.Request.Context(), query.GetRoleQuery{ID: id})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": view})
}

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
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteRole handles DELETE /roles/:id.
func (h *Handler) DeleteRole(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	err = h.bc.DeleteRole.Handle(ctx.Request.Context(), command.DeleteRoleCommand{ID: id})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// --- Permissions ---

// CreatePermission handles POST /permissions.
func (h *Handler) CreatePermission(ctx *gin.Context) {
	var req CreatePermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	err := h.bc.CreatePermission.Handle(ctx.Request.Context(), command.CreatePermissionCommand{
		Name:        req.Name,
		ParentID:    req.ParentID,
		Description: req.Description,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

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

// DeletePermission handles DELETE /permissions/:id.
func (h *Handler) DeletePermission(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	err = h.bc.DeletePermission.Handle(ctx.Request.Context(), command.DeletePermissionCommand{ID: id})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// --- Policies ---

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
		ID:         id,
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

// DeletePolicy handles DELETE /policies/:id.
func (h *Handler) DeletePolicy(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	err = h.bc.DeletePolicy.Handle(ctx.Request.Context(), command.DeletePolicyCommand{ID: id})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// TogglePolicy handles POST /policies/:id/toggle.
func (h *Handler) TogglePolicy(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	err = h.bc.TogglePolicy.Handle(ctx.Request.Context(), command.TogglePolicyCommand{ID: id})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// --- Scopes ---

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
		RoleID:       roleID,
		PermissionID: req.PermissionID,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

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
		PermissionID: permID,
		Path:         req.Path,
		Method:       req.Method,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
