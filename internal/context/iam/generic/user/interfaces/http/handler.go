package http

import (
	"net/http"

	"gct/internal/context/iam/generic/user"
	"gct/internal/context/iam/generic/user/application/command"
	"gct/internal/context/iam/generic/user/application/query"
	userentity "gct/internal/context/iam/generic/user/domain/entity"
	"gct/internal/kernel/infrastructure/httpx"
	"gct/internal/kernel/infrastructure/httpx/response"
	"gct/internal/kernel/infrastructure/logger"
	"gct/internal/kernel/infrastructure/security/fingerprint"

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

// @Summary Create a user
// @Description Create a new user account
// @Tags Users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User data"
// @Success 201 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /users [post]
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

// @Summary List users
// @Description Get a paginated list of users with optional filters
// @Tags Users
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Param phone query string false "Filter by phone"
// @Param email query string false "Filter by email"
// @Param active query bool false "Filter by active status"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /users [get]
// List handles GET /users.
func (h *Handler) List(ctx *gin.Context) {
	pg, err := httpx.GetPagination(ctx)
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParamIsInvalid, http.StatusBadRequest)
		return
	}

	filter := userentity.UsersFilter{
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

// @Summary Get a user
// @Description Get user details by ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /users/{id} [get]
// Get handles GET /users/:id.
func (h *Handler) Get(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	view, err := h.bc.GetUser.Handle(ctx.Request.Context(), query.GetUserQuery{ID: userentity.UserID(id)})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": view})
}

// @Summary Update a user
// @Description Update user details by ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param request body UpdateUserRequest true "User update data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /users/{id} [patch]
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
		ID:         userentity.UserID(id),
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

// @Summary Delete a user
// @Description Delete a user by ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /users/{id} [delete]
// Delete handles DELETE /users/:id.
func (h *Handler) Delete(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	err = h.bc.DeleteUser.Handle(ctx.Request.Context(), command.DeleteUserCommand{ID: userentity.UserID(id)})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// @Summary Approve a user
// @Description Approve a pending user account by ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /users/{id}/approve [post]
// Approve handles POST /users/:id/approve.
func (h *Handler) Approve(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.RespondWithError(ctx, httpx.ErrParsingUUID, http.StatusBadRequest)
		return
	}

	err = h.bc.ApproveUser.Handle(ctx.Request.Context(), command.ApproveUserCommand{ID: userentity.UserID(id)})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// @Summary Change user role
// @Description Change the role of a user by ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID (UUID)"
// @Param request body ChangeRoleRequest true "Role data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /users/{id}/role [post]
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
		UserID: userentity.UserID(id),
		RoleID: req.RoleID,
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// @Summary Bulk action on users
// @Description Perform a bulk action on multiple users
// @Tags Users
// @Accept json
// @Produce json
// @Param request body BulkActionRequest true "Bulk action data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /users/bulk-action [post]
// BulkAction handles POST /users/bulk-action.
func (h *Handler) BulkAction(ctx *gin.Context) {
	var req BulkActionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	ids := make([]userentity.UserID, len(req.IDs))
	for i, id := range req.IDs {
		ids[i] = userentity.UserID(id)
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

// @Summary Sign in
// @Description Authenticate a user and return tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body SignInRequest true "Sign-in credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/sign-in [post]
// SignIn handles POST /auth/sign-in.
func (h *Handler) SignIn(ctx *gin.Context) {
	var req SignInRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	apiKey := httpx.GetAPIKey(ctx)
	if apiKey == "" {
		response.RespondWithError(ctx, httpx.ErrUnAuth, http.StatusUnauthorized)
		return
	}

	fp := fingerprint.Compute(
		ctx.Request.UserAgent(),
		ctx.Request.Header.Get("Accept-Language"),
		ctx.Request.Header.Get("Sec-CH-UA"),
	)

	result, err := h.bc.SignIn.Handle(ctx.Request.Context(), command.SignInCommand{
		Login:             req.Login,
		Password:          req.Password,
		DeviceType:        req.DeviceType,
		IP:                ctx.ClientIP(),
		UserAgent:         ctx.GetHeader("User-Agent"),
		APIKey:            apiKey,
		DeviceFingerprint: fp,
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

// @Summary Sign up
// @Description Register a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body SignUpRequest true "Sign-up data"
// @Success 201 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/sign-up [post]
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

// @Summary Sign out
// @Description End a user session
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body SignOutRequest true "Sign-out data"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /auth/sign-out [post]
// SignOut handles POST /auth/sign-out.
func (h *Handler) SignOut(ctx *gin.Context) {
	var req SignOutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.RespondWithError(ctx, err, http.StatusBadRequest)
		return
	}

	err := h.bc.SignOut.Handle(ctx.Request.Context(), command.SignOutCommand{
		UserID:    userentity.UserID(req.UserID),
		SessionID: userentity.SessionID(req.SessionID),
		IP:        ctx.ClientIP(),
		UserAgent: ctx.Request.UserAgent(),
	})
	if err != nil {
		response.HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true})
}
