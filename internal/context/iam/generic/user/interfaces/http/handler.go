package http

import (
	"net/http"

	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/context/iam/generic/user"
	"gct/internal/context/iam/generic/user/application/command"
	"gct/internal/context/iam/generic/user/application/query"
	userdomain "gct/internal/context/iam/generic/user/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler holds dependencies for User HTTP handlers.
type Handler struct {
	bc *user.BoundedContext
	l  logger.Log
}

// NewHandler creates a new User HTTP handler.
func NewHandler(bc *user.BoundedContext, l logger.Log) *Handler {
	return &Handler{bc: bc, l: l}
}

// Create handles POST /users.
func (h *Handler) Create(ctx *gin.Context) {
	var req CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	err := h.bc.CreateUser.Handle(ctx.Request.Context(), command.CreateUserCommand{
		Phone:      req.Phone,
		Password:   req.Password,
		Email:      req.Email,
		Username:   req.Username,
		RoleID:     req.RoleID,
		Attributes: req.Attributes,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"success": true})
}

// List handles GET /users.
func (h *Handler) List(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	filter := userdomain.UsersFilter{
		Pagination: &pg,
	}

	if phone := ctx.Query("phone"); phone != "" {
		filter.Phone = &phone
	}
	if email := ctx.Query("email"); email != "" {
		filter.Email = &email
	}
	if activeStr := ctx.Query("active"); activeStr != "" {
		active := activeStr == "true"
		filter.Active = &active
	}

	result, err := h.bc.ListUsers.Handle(ctx.Request.Context(), query.ListUsersQuery{
		Filter: filter,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  result.Users,
		"total": result.Total,
	})
}

// Get handles GET /users/:id.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	view, err := h.bc.GetUser.Handle(ctx.Request.Context(), query.GetUserQuery{ID: userdomain.UserID(id)})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": view})
}

// Update handles PATCH /users/:id.
func (h *Handler) Update(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	var req UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	err = h.bc.UpdateUser.Handle(ctx.Request.Context(), command.UpdateUserCommand{
		ID:         userdomain.UserID(id),
		Email:      req.Email,
		Username:   req.Username,
		Attributes: req.Attributes,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// Delete handles DELETE /users/:id.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	err = h.bc.DeleteUser.Handle(ctx.Request.Context(), command.DeleteUserCommand{ID: userdomain.UserID(id)})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// Approve handles POST /users/:id/approve.
func (h *Handler) Approve(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	err = h.bc.ApproveUser.Handle(ctx.Request.Context(), command.ApproveUserCommand{ID: userdomain.UserID(id)})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// ChangeRole handles POST /users/:id/role.
func (h *Handler) ChangeRole(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	var req ChangeRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	err = h.bc.ChangeRole.Handle(ctx.Request.Context(), command.ChangeRoleCommand{
		UserID: userdomain.UserID(id),
		RoleID: req.RoleID,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// BulkAction handles POST /users/bulk-action.
func (h *Handler) BulkAction(ctx *gin.Context) {
	var req BulkActionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	ids := make([]userdomain.UserID, len(req.IDs))
	for i, id := range req.IDs {
		ids[i] = userdomain.UserID(id)
	}
	err := h.bc.BulkAction.Handle(ctx.Request.Context(), command.BulkActionCommand{
		IDs:    ids,
		Action: req.Action,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// SignIn handles POST /auth/sign-in.
func (h *Handler) SignIn(ctx *gin.Context) {
	var req SignInRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	result, err := h.bc.SignIn.Handle(ctx.Request.Context(), command.SignInCommand{
		Login:      req.Login,
		Password:   req.Password,
		DeviceType: req.DeviceType,
		IP:         ctx.ClientIP(),
		UserAgent:  ctx.GetHeader("User-Agent"),
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"user_id":       result.UserID,
			"session_id":    result.SessionID,
			"access_token":  result.AccessToken,
			"refresh_token": result.RefreshToken,
		},
	})
}

// SignUp handles POST /auth/sign-up.
func (h *Handler) SignUp(ctx *gin.Context) {
	var req SignUpRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	err := h.bc.SignUp.Handle(ctx.Request.Context(), command.SignUpCommand{
		Phone:    req.Phone,
		Password: req.Password,
		Username: req.Username,
		Email:    req.Email,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success"})
}

// SignOut handles POST /auth/sign-out.
func (h *Handler) SignOut(ctx *gin.Context) {
	var req SignOutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	err := h.bc.SignOut.Handle(ctx.Request.Context(), command.SignOutCommand{
		UserID:    userdomain.UserID(req.UserID),
		SessionID: userdomain.SessionID(req.SessionID),
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
