package admin

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"gct/config"
	"gct/consts"
	"gct/internal/controller/restapi/cookie"
	auth "gct/internal/controller/restapi/middleware/auth"
	"gct/internal/domain"
	"gct/internal/usecase"
	apperrors "gct/pkg/errors"
	"gct/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	uc  *usecase.UseCase
	cfg *config.Config
	l   logger.Log
}

func New(uc *usecase.UseCase, cfg *config.Config, l logger.Log) *Handler {
	return &Handler{
		uc:  uc,
		cfg: cfg,
		l:   l,
	}
}

// Helper function to convert string to pointer
func strPtr(s string) *string {
	return &s
}

func (h *Handler) Register(r *gin.RouterGroup, authmw *auth.AuthMiddleware) {
	g := r.Group("/admin")
	g.GET("", h.AdminRoot)
	g.GET("/login", h.Login)
	g.POST("/login", h.LoginPost)
	g.GET("/logout", h.Logout)
	g.POST("/logout", h.Logout)

	// Setup Routes
	g.GET("/setup", h.Setup)
	g.POST("/setup", h.SetupPost)

	// Register Routes
	g.GET("/register", h.RegisterPage)
	g.POST("/register", h.RegisterPost)

	protected := g.Group("/")
	protected.Use(authmw.AuthWeb)

	protected.GET("/dashboard", h.Dashboard)
	protected.GET("/users", h.Users)
	protected.GET("/users/create", h.CreateUser)
	protected.POST("/users/create", h.CreateUserPost)
	protected.GET("/users/:id", h.UserDetail)
	protected.GET("/users/:id/edit", h.EditUser)
	protected.POST("/users/:id/edit", h.UpdateUserPost)
	protected.POST("/users/bulk-action", h.BulkUsersAction)
	protected.GET("/approvals", h.Approvals)
	protected.POST("/users/:id/approve", h.ApproveUser)
	protected.GET("/sessions", h.Sessions)
	protected.GET("/sessions/:id", h.SessionDetail)
	protected.POST("/sessions/:id/revoke", h.RevokeSession)
	protected.GET("/rbac/roles", h.Roles)
	protected.GET("/rbac/permissions", h.Permissions)
	protected.GET("/rbac/scopes", h.Scopes)
	protected.GET("/abac/policies", h.Policies)
	protected.GET("/system-errors", h.SystemErrors)
	protected.GET("/functions", h.FunctionMetrics)
	protected.GET("/audit/logs", h.AuditLogs)
	protected.GET("/audit/history", h.EndpointHistory)
	protected.GET("/linter", h.Linter)
	protected.POST("/linter/run", h.RunLinter)
	// Asynq Monitor
	mon := h.NewAsynqMonitor()
	protected.Any("/asynq/*filepath", gin.WrapH(mon))
	protected.GET("/asynq", gin.WrapH(mon))

	protected.GET("/settings", h.Settings)
	protected.POST("/settings/:id", h.UpdateSetting)
	protected.GET("/api/stats", h.DashboardStats)

	// Database
	protected.GET("/database/monitoring", h.DatabaseMonitoring)
	protected.GET("/database/tables", h.DatabaseTables)
	protected.GET("/database/table/:name", h.TableData)
	protected.GET("/database/api/table/:name/data", h.GetTableDataAPI)
	protected.POST("/database/api/table/:name/record", h.CreateRecord)
	protected.GET("/database/api/tables", h.GetTablesAPI)
	protected.PUT("/database/api/table/:name/record/:id", h.UpdateRecord)
	protected.DELETE("/database/api/table/:name/record/:id", h.DeleteRecord)
	protected.GET("/database/sql-editor", h.SQLEditor)
	protected.POST("/database/sql-editor/execute", h.ExecuteSQL)

	protected.GET("/profile", h.Profile)
	protected.POST("/profile", h.ProfilePost)
	protected.POST("/users/:id/:action", h.UserAction)

	// Section Overviews
	protected.GET("/section/users", h.UsersOverview)
	protected.GET("/section/access", h.AccessControlOverview)
	protected.GET("/section/database", h.DatabaseOverview)
	protected.GET("/section/monitoring", h.MonitoringOverview)
	protected.GET("/section/tools", h.ToolsOverview)
}

// UsersOverview - Overview page for Users Section
func (h *Handler) UsersOverview(ctx *gin.Context) {
	ctxReq := ctx.Request.Context()
	limit := &domain.Pagination{Limit: 1}

	_, usersCount, _ := h.uc.User.Client().Gets(ctxReq, &domain.UsersFilter{Pagination: limit})
	_, sessionsCount, _ := h.uc.User.Session().Gets(ctxReq, &domain.SessionsFilter{Pagination: limit})

	// Pending approvals
	notApproved := false
	_, pendingCount, _ := h.uc.User.Client().Gets(ctxReq, &domain.UsersFilter{
		Pagination: limit,
		UserFilter: domain.UserFilter{IsApproved: &notApproved},
	})

	data := map[string]any{
		"UsersCount":    usersCount,
		"SessionsCount": sessionsCount,
		"PendingCount":  pendingCount,
	}
	h.servePage(ctx, "section_users.html", "User Management", "users_overview", data)
}

// AccessControlOverview - Overview page for Access Control
func (h *Handler) AccessControlOverview(ctx *gin.Context) {
	ctxReq := ctx.Request.Context()
	limit := &domain.Pagination{Limit: 1}

	_, rolesCount, _ := h.uc.Authz.Role().Gets(ctxReq, &domain.RolesFilter{Pagination: limit})
	_, permsCount, _ := h.uc.Authz.Permission().Gets(ctxReq, &domain.PermissionsFilter{Pagination: limit})
	_, scopesCount, _ := h.uc.Authz.Scope().Gets(ctxReq, &domain.ScopesFilter{Pagination: limit})
	_, policiesCount, _ := h.uc.Authz.Policy().Gets(ctxReq, &domain.PoliciesFilter{Pagination: limit})

	data := map[string]any{
		"RolesCount":    rolesCount,
		"PermsCount":    permsCount,
		"ScopesCount":   scopesCount,
		"PoliciesCount": policiesCount,
	}
	h.servePage(ctx, "section_access.html", "Access Control", "access_overview", data)
}

// DatabaseOverview - Overview page for Database
func (h *Handler) DatabaseOverview(ctx *gin.Context) {
	// Mock stats or fetch real ones if available
	data := map[string]any{
		"TablesCount": 12, // Placeholder or fetch actual
		"ActiveConn":  5,  // Placeholder
	}
	h.servePage(ctx, "section_database.html", "Database Management", "database_overview", data)
}

// MonitoringOverview - Overview for Monitoring
func (h *Handler) MonitoringOverview(ctx *gin.Context) {
	ctxReq := ctx.Request.Context()
	limit := &domain.Pagination{Limit: 1}

	_, auditCount, _ := h.uc.Audit.Log().Gets(ctxReq, &domain.AuditLogsFilter{Pagination: limit})

	data := map[string]any{
		"AuditCount": auditCount,
		"ErrorCount": 0, // Placeholder
	}
	h.servePage(ctx, "section_monitoring.html", "System Monitoring", "monitoring_overview", data)
}

// ToolsOverview - Overview for Tools
func (h *Handler) ToolsOverview(ctx *gin.Context) {
	data := map[string]any{
		"LinterStatus": "Active",
	}
	h.servePage(ctx, "section_tools.html", "Tools & Settings", "tools_overview", data)
}

func (h *Handler) AdminRoot(ctx *gin.Context) {
	// Check if user is already authenticated
	accessToken := cookie.GetCookie(ctx, consts.COOKIE_ACCESS_TOKEN)
	if accessToken != "" {
		// User has a token, redirect to dashboard
		ctx.Redirect(http.StatusFound, "/admin/dashboard")
		return
	}
	// Not authenticated, redirect to login
	ctx.Redirect(http.StatusFound, "/admin/login")
}

func (h *Handler) Login(ctx *gin.Context) {
	// Check if user is already authenticated
	accessToken := cookie.GetCookie(ctx, consts.COOKIE_ACCESS_TOKEN)
	if accessToken != "" {
		// User has a token, redirect to dashboard
		ctx.Redirect(http.StatusFound, "/admin/dashboard")
		return
	}

	// Check if any users exist
	users, _, err := h.uc.User.Client().Gets(ctx.Request.Context(), &domain.UsersFilter{
		Pagination: &domain.Pagination{Limit: 1, Offset: 0},
	})
	if err == nil && len(users) == 0 {
		h.l.Infow("No users found, redirecting to register")
		ctx.Redirect(http.StatusFound, "/admin/register")
		return
	}

	tmpl, err := template.ParseFiles("internal/web/admin/templates/pages/login.html")
	if err != nil {
		h.l.Errorw("failed to parse login template", "error", err)
		ctx.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if err := tmpl.Execute(ctx.Writer, map[string]any{
		"Error": ctx.Query("error"),
		"CSRF":  ctx.GetString("csrf"),
	}); err != nil {
		h.l.Errorw("failed to execute login template", "error", err)
	}
}

// LoginPost (POST) - Authenticate User
func (h *Handler) LoginPost(ctx *gin.Context) {
	email := ctx.PostForm("email")
	password := ctx.PostForm("password")

	// IP/UA
	ip := ctx.ClientIP()
	ua := ctx.Request.UserAgent()

	// SignIn Usecase now supports Email
	signInInput := &domain.SignInIn{
		Login:    strPtr(email),
		Password: strPtr(password),
	}
	signInInput.Session.IP = ip
	signInInput.Session.UserAgent = ua
	res, err := h.uc.User.Client().SignIn(ctx.Request.Context(), signInInput)
	if err != nil {
		h.l.Warnw("login failed", "email", email, "error", err)
		ctx.Redirect(http.StatusFound, "/admin/login?error=Invalid+credentials+or+inactive+account")
		return
	}

	// Set Cookies
	isSecure := ctx.Request.TLS != nil || ctx.Request.Header.Get("X-Forwarded-Proto") == "https"

	remember := ctx.PostForm("remember")
	refreshMaxAge := 0 // Session cookie by default
	if remember == "on" {
		refreshMaxAge = int(h.cfg.JWT.RefreshTTL.Seconds())
	}

	ctx.SetCookie(consts.COOKIE_ACCESS_TOKEN, res.AccessToken, int(h.cfg.JWT.AccessTTL.Seconds()), "/", "", isSecure, true)
	ctx.SetCookie(consts.COOKIE_REFRESH_TOKEN, res.RefreshToken, refreshMaxAge, "/", "", isSecure, true)

	ctx.Redirect(http.StatusFound, "/admin/dashboard")
}

func (h *Handler) Logout(ctx *gin.Context) {
	cookie.ExpireCookies(ctx, h.cfg.Cookie, consts.COOKIE_ACCESS_TOKEN, consts.COOKIE_REFRESH_TOKEN)
	ctx.Redirect(http.StatusFound, "/admin/login")
}

// PageData holds common data for all admin pages
type PageData struct {
	Title          string
	ActivePage     string
	ActiveMenu     string
	User           *domain.User
	Can            map[string]bool
	TracingEnabled bool
	JaegerURL      string
	RoleName       string
	Data           any
}

// servePage prepares common data and renders the template
func (h *Handler) servePage(ctx *gin.Context, tmplName, title, activeMenu string, data any) {
	var user *domain.User
	if u, exists := ctx.Get("user"); exists {
		if usr, ok := u.(*domain.User); ok {
			user = usr
		}
	}

	// Fetch Role Name from Context (injected by AuthWeb)
	roleName := ctx.GetString(consts.CtxRoleTitle)

	// If data is nil, initialize it
	if data == nil {
		data = make(map[string]any)
	}

	pageData := PageData{
		Title:          title,
		ActiveMenu:     activeMenu,
		User:           user,
		TracingEnabled: h.cfg.Tracing.Enabled,
		JaegerURL:      h.cfg.Tracing.Jaeger.URL,
		RoleName:       roleName,
		Data:           data,
	}

	h.render(ctx, tmplName, pageData)
}

func (h *Handler) render(ctx *gin.Context, tmplName string, data PageData) {
	files := []string{
		"internal/web/admin/templates/layout/base.html",
		"internal/web/admin/templates/layout/header.html",
		"internal/web/admin/templates/layout/sidebar.html",
		"internal/web/admin/templates/layout/pagination.html",
		filepath.Join("internal/web/admin/templates/pages", tmplName),
	}

	// Include all partials (top level and one level deep)
	partials, _ := filepath.Glob("internal/web/admin/templates/partials/*.html")
	if nestedPartials, err := filepath.Glob("internal/web/admin/templates/partials/*/*.html"); err == nil {
		partials = append(partials, nestedPartials...)
	}
	files = append(files, partials...)

	// Clone the global func map
	funcs := template.FuncMap{}
	for k, v := range templateFuncs {
		funcs[k] = v
	}

	// Add context-aware overrides if needed
	funcs["currUserEmail"] = func() string {
		// Use the global helper but pass the current user from context data
		f := templateFuncs["currUserEmail"].(func(*domain.User) string)
		return f(data.User)
	}

	tmpl, err := template.New("base").Funcs(funcs).ParseFiles(files...)
	if err != nil {
		h.l.Errorw("failed to parse templates", "error", err)
		ctx.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if err := tmpl.Execute(ctx.Writer, data); err != nil {
		h.l.Errorw("failed to execute template", "error", err)
	}
}

func (h *Handler) bindPagination(ctx *gin.Context) *domain.Pagination {
	page, _ := strconv.ParseInt(ctx.DefaultQuery("page", "1"), 10, 64)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.ParseInt(ctx.DefaultQuery("limit", "10"), 10, 64)
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return &domain.Pagination{
		Limit:     limit,
		Offset:    (page - 1) * limit,
		SortBy:    ctx.Query("sort_by"),
		SortOrder: ctx.DefaultQuery("sort_order", "DESC"),
	}
}

func (h *Handler) Dashboard(ctx *gin.Context) {
	ctxReq := ctx.Request.Context()
	limit := &domain.Pagination{Limit: 1}

	// Fetch counts
	_, usersCount, _ := h.uc.User.Client().Gets(ctxReq, &domain.UsersFilter{Pagination: limit})
	_, sessionsCount, _ := h.uc.User.Session().Gets(ctxReq, &domain.SessionsFilter{Pagination: limit})
	_, rolesCount, _ := h.uc.Authz.Role().Gets(ctxReq, &domain.RolesFilter{Pagination: limit})
	_, permsCount, _ := h.uc.Authz.Permission().Gets(ctxReq, &domain.PermissionsFilter{Pagination: limit})
	_, scopesCount, _ := h.uc.Authz.Scope().Gets(ctxReq, &domain.ScopesFilter{Pagination: limit})
	_, policiesCount, _ := h.uc.Authz.Policy().Gets(ctxReq, &domain.PoliciesFilter{Pagination: limit})
	_, auditCount, _ := h.uc.Audit.Log().Gets(ctxReq, &domain.AuditLogsFilter{Pagination: limit})

	data := map[string]any{
		"CurrentDate":   time.Now().Format("Monday, January 2, 2006"),
		"UsersCount":    usersCount,
		"SessionsCount": sessionsCount,
		"RolesCount":    rolesCount,
		"PermsCount":    permsCount,
		"ScopesCount":   scopesCount,
		"PoliciesCount": policiesCount,
		"AuditCount":    auditCount,
	}
	h.servePage(ctx, "dashboard.html", "Dashboard", "dashboard", data)
}

func (h *Handler) DashboardStats(ctx *gin.Context) {
	ctxReq := ctx.Request.Context()
	limit := &domain.Pagination{Limit: 1}

	// Fetch counts in parallel logic could be used, but sequential is fine for now on small scale.
	// 1. Users
	_, usersCount, _ := h.uc.User.Client().Gets(ctxReq, &domain.UsersFilter{Pagination: limit})
	// 2. Sessions
	_, sessionsCount, _ := h.uc.User.Session().Gets(ctxReq, &domain.SessionsFilter{Pagination: limit})
	// 3. Roles
	_, rolesCount, _ := h.uc.Authz.Role().Gets(ctxReq, &domain.RolesFilter{Pagination: limit})
	// 4. Permissions
	_, permsCount, _ := h.uc.Authz.Permission().Gets(ctxReq, &domain.PermissionsFilter{Pagination: limit})
	// 5. Scopes
	_, scopesCount, _ := h.uc.Authz.Scope().Gets(ctxReq, &domain.ScopesFilter{Pagination: limit})
	// 6. Policies
	_, policiesCount, _ := h.uc.Authz.Policy().Gets(ctxReq, &domain.PoliciesFilter{Pagination: limit})

	ctx.JSON(http.StatusOK, gin.H{
		"users_count":    usersCount,
		"sessions_count": sessionsCount,
		"roles_count":    rolesCount,
		"perms_count":    permsCount,
		"scopes_count":   scopesCount,
		"policies_count": policiesCount,
	})
}

func (h *Handler) Users(ctx *gin.Context) {
	pagination := h.bindPagination(ctx)
	filter := &domain.UsersFilter{Pagination: pagination}

	if email := ctx.Query("email"); email != "" {
		filter.Email = &email
	}
	if role := ctx.Query("role_id"); role != "" {
		if uid, err := uuid.Parse(role); err == nil {
			filter.RoleID = &uid
		}
	}
	if active := ctx.Query("active"); active != "" {
		b := active == "true"
		filter.Active = &b
	}

	users, count, err := h.uc.User.Client().Gets(ctx.Request.Context(), filter)
	if err != nil {
		h.l.Errorw("failed to fetch users", "error", err)
	}
	pagination.Total = int64(count)

	h.servePage(ctx, "users/list.html", "Users", "users", map[string]any{
		"Users":       users,
		"Pagination":  pagination,
		"Filter":      filter,
		"QueryParams": ctx.Request.URL.Query(),
	})
}

func (h *Handler) Sessions(ctx *gin.Context) {
	pagination := h.bindPagination(ctx)

	// Default sort by created_at DESC if not specified
	if pagination.SortBy == "" {
		pagination.SortBy = "created_at"
	}

	filter := &domain.SessionsFilter{Pagination: pagination}

	if uid := ctx.Query("user_id"); uid != "" {
		if id, err := uuid.Parse(uid); err == nil {
			filter.UserID = &id
		}
	}

	if revoked := ctx.Query("revoked"); revoked != "" {
		b := revoked == "true"
		filter.Revoked = &b
	}

	sessions, count, err := h.uc.User.Session().Gets(ctx.Request.Context(), filter)
	if err != nil {
		h.l.Errorw("failed to fetch sessions", "error", err)
	}
	pagination.Total = int64(count)

	h.servePage(ctx, "sessions/list.html", "Sessions", "sessions", map[string]any{
		"Sessions":    sessions,
		"Pagination":  pagination,
		"QueryParams": ctx.Request.URL.Query(),
	})
}

func (h *Handler) SessionDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.l.Errorw("invalid session id", "error", err)
		ctx.Redirect(http.StatusFound, "/admin/sessions")
		return
	}

	// Use Gets with ID filter instead of Get to avoid expiration check logic in Get
	// We want to see expired sessions too in admin panel
	limit := &domain.Pagination{Limit: 1}
	filter := &domain.SessionsFilter{
		SessionFilter: domain.SessionFilter{ID: &id},
		Pagination:    limit,
	}

	sessions, count, err := h.uc.User.Session().Gets(ctx.Request.Context(), filter)
	if err != nil || count == 0 {
		h.l.Errorw("failed to fetch session", "error", err)
		ctx.Redirect(http.StatusFound, "/admin/sessions")
		return
	}

	h.servePage(ctx, "sessions/detail.html", "Session Details", "sessions", map[string]any{
		"Session": sessions[0],
	})
}

func (h *Handler) RevokeSession(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	revoke := true
	filter := &domain.SessionFilter{
		ID:      &id,
		Revoked: &revoke,
	}

	err = h.uc.User.Session().Revoke(ctx.Request.Context(), filter)
	if err != nil {
		h.l.Errorw("failed to revoke session", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke session"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Session revoked successfully"})
}

func (h *Handler) Roles(ctx *gin.Context) {
	pagination := h.bindPagination(ctx)
	filter := &domain.RolesFilter{
		Pagination: pagination,
	}

	if name := ctx.Query("name"); name != "" {
		filter.Name = &name
	}

	roles, count, err := h.uc.Authz.Role().Gets(ctx.Request.Context(), filter)
	if err != nil {
		h.l.Errorw("failed to fetch roles", "error", err)
	}
	pagination.Total = int64(count)

	h.servePage(ctx, "rbac/roles.html", "Roles", "roles", map[string]any{
		"Roles":       roles,
		"Pagination":  pagination,
		"Filter":      filter,
		"QueryParams": ctx.Request.URL.Query(),
	})
}

func (h *Handler) Permissions(ctx *gin.Context) {
	pagination := h.bindPagination(ctx)
	filter := &domain.PermissionsFilter{
		Pagination: pagination,
	}

	if name := ctx.Query("name"); name != "" {
		filter.Name = &name
	}

	perms, count, err := h.uc.Authz.Permission().Gets(ctx.Request.Context(), filter)
	if err != nil {
		h.l.Errorw("failed to fetch permissions", "error", err)
	}
	pagination.Total = int64(count)

	h.servePage(ctx, "rbac/permissions.html", "Permissions", "permissions", map[string]any{
		"Permissions": perms,
		"Pagination":  pagination,
		"Filter":      filter,
		"QueryParams": ctx.Request.URL.Query(),
	})
}

func (h *Handler) Scopes(ctx *gin.Context) {
	pagination := h.bindPagination(ctx)
	filter := &domain.ScopesFilter{
		Pagination: pagination,
	}

	if path := ctx.Query("path"); path != "" {
		filter.Path = &path
	}
	if method := ctx.Query("method"); method != "" {
		filter.Method = &method
	}

	scopes, count, err := h.uc.Authz.Scope().Gets(ctx.Request.Context(), filter)
	if err != nil {
		h.l.Errorw("failed to fetch scopes", "error", err)
	}
	pagination.Total = int64(count)

	h.servePage(ctx, "rbac/scopes.html", "Scopes", "scopes", map[string]any{
		"Scopes":      scopes,
		"Pagination":  pagination,
		"Filter":      filter,
		"QueryParams": ctx.Request.URL.Query(),
	})
}

func (h *Handler) Policies(ctx *gin.Context) {
	pagination := h.bindPagination(ctx)
	filter := &domain.PoliciesFilter{
		Pagination: pagination,
	}

	if active := ctx.Query("active"); active != "" {
		b := active == "true"
		filter.Active = &b
	}

	policies, count, err := h.uc.Authz.Policy().Gets(ctx.Request.Context(), filter)
	if err != nil {
		h.l.Errorw("failed to fetch policies", "error", err)
	}
	pagination.Total = int64(count)

	h.servePage(ctx, "abac/policies.html", "Policies", "policies", map[string]any{
		"Policies":    policies,
		"Pagination":  pagination,
		"Filter":      filter,
		"QueryParams": ctx.Request.URL.Query(),
	})
}

func (h *Handler) Linter(ctx *gin.Context) {
	h.servePage(ctx, "linter.html", "Code Linter", "linter", nil)
}

func (h *Handler) Settings(ctx *gin.Context) {
	// Fetch all site settings
	settings, _, err := h.uc.SiteSetting.Gets(ctx.Request.Context(), &domain.SiteSettingsFilter{
		Pagination: &domain.Pagination{Limit: 100},
	})
	if err != nil {
		h.l.Errorw("failed to fetch site settings", "error", err)
	}

	// Organize settings by category
	settingsByCategory := make(map[string][]*domain.SiteSetting)
	for _, setting := range settings {
		settingsByCategory[setting.Category] = append(settingsByCategory[setting.Category], setting)
	}

	h.servePage(ctx, "settings.html", "Settings", "settings", map[string]any{
		"Config":             h.cfg,
		"Settings":           settings,
		"SettingsByCategory": settingsByCategory,
		"CSRF":               ctx.GetString("csrf"),
	})
}

func (h *Handler) UpdateSetting(ctx *gin.Context) {
	settingID := ctx.Param("id")
	id, err := uuid.Parse(settingID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid setting ID"})
		return
	}

	var req struct {
		Value string `json:"value"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request"})
		return
	}

	// Get existing setting
	setting, err := h.uc.SiteSetting.Get(ctx.Request.Context(), &domain.SiteSettingFilter{ID: &id})
	if err != nil {
		h.l.Errorw("failed to get setting", "id", id, "error", err)
		ctx.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Setting not found"})
		return
	}

	// Update value
	setting.Value = req.Value
	err = h.uc.SiteSetting.Update(ctx.Request.Context(), setting)
	if err != nil {
		h.l.Errorw("failed to update setting", "id", id, "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update setting"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Setting updated successfully"})
}

func (h *Handler) Setup(ctx *gin.Context) {
	// Check if users already exist to prevent accessing setup again
	_, count, err := h.uc.User.Client().Gets(ctx.Request.Context(), &domain.UsersFilter{
		Pagination: &domain.Pagination{Limit: 1},
	})
	if err == nil && count > 0 {
		ctx.Redirect(http.StatusFound, "/admin/login")
		return
	}

	tmpl, err := template.ParseFiles("internal/web/admin/templates/pages/setup.html")
	if err != nil {
		h.l.Errorw("failed to parse setup template", "error", err)
		ctx.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if err := tmpl.Execute(ctx.Writer, map[string]any{"Error": ctx.Query("error"), "CSRF": ctx.GetString("csrf")}); err != nil {
		h.l.Errorw("failed to execute setup template", "error", err)
	}
}

func (h *Handler) SetupPost(ctx *gin.Context) {
	email := ctx.PostForm("email")
	password := ctx.PostForm("password")
	phone := ctx.PostForm("phone") // Optional or required

	if email == "" || password == "" {
		ctx.Redirect(http.StatusFound, "/admin/setup?error=Email+and+Password+required")
		return
	}

	// 1. Get Admin Role
	// Accessing Repo directly via Usecase (UseCase -> Repo -> Postgres -> Authz -> Role)
	name := "admin"
	role, err := h.uc.Repo.Persistent.Postgres.Authz.Role.Get(ctx.Request.Context(), &domain.RoleFilter{Name: &name})
	if err != nil {
		h.l.Errorw("setup failed: admin role not found", "error", err)
		ctx.Redirect(http.StatusFound, "/admin/setup?error=System+configuration+error+(role+missing)")
		return
	}

	// 2. Create User
	u := domain.NewUser()
	u.Email = &email
	if phone != "" {
		u.Phone = &phone
	}
	u.RoleID = &role.ID
	u.Active = true
	u.IsApproved = true // Auto-approve super admin

	if err := u.SetPassword(password); err != nil {
		ctx.Redirect(http.StatusFound, "/admin/setup?error=Invalid+password")
		return
	}

	err = h.uc.User.Client().Create(ctx.Request.Context(), u)
	if err != nil {
		h.l.Errorw("setup failed: create user", "error", err)
		ctx.Redirect(http.StatusFound, "/admin/setup?error=Failed+to+create+user:+"+err.Error())
		return
	}

	ctx.Redirect(http.StatusFound, "/admin/login?success=Setup+Complete")
}

// RegisterPage (GET) - Show Sign Up Form
func (h *Handler) RegisterPage(ctx *gin.Context) {
	tmpl, err := template.ParseFiles("internal/web/admin/templates/pages/register.html")
	if err != nil {
		h.l.Errorw("failed to parse register template", "error", err)
		ctx.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if err := tmpl.Execute(ctx.Writer, map[string]any{"Error": ctx.Query("error"), "CSRF": ctx.GetString("csrf")}); err != nil {
		h.l.Errorw("failed to execute register template", "error", err)
	}
}

// RegisterPost (POST) - Process Sign Up
func (h *Handler) RegisterPost(ctx *gin.Context) {
	email := ctx.PostForm("email")
	password := ctx.PostForm("password")
	phone := ctx.PostForm("phone")

	// Basic validation
	if email == "" || password == "" {
		ctx.Redirect(http.StatusFound, "/admin/register?error=Email+and+Password+are+required")
		return
	}

	u := domain.NewUser()
	u.Email = &email
	u.Phone = &phone // Phone optional? Form says required.
	if phone != "" {
		u.Phone = &phone
	}

	if err := u.SetPassword(password); err != nil {
		ctx.Redirect(http.StatusFound, "/admin/register?error=Invalid+Password")
		return
	}

	// Check if this is the first user in the system
	existingUsers, _, err := h.uc.User.Client().Gets(ctx.Request.Context(), &domain.UsersFilter{
		Pagination: &domain.Pagination{Limit: 1, Offset: 0},
	})

	isFirstUser := err == nil && len(existingUsers) == 0

	if isFirstUser {
		// First user gets super_admin role and auto-approval
		roleFilter := &domain.RoleFilter{Name: strPtr("super_admin")}
		adminRole, roleErr := h.uc.Repo.Persistent.Postgres.Authz.Role.Get(ctx.Request.Context(), roleFilter)
		if roleErr != nil {
			h.l.Errorw("failed to get super_admin role", "error", roleErr)
			ctx.Redirect(http.StatusFound, "/admin/register?error=System+configuration+error")
			return
		}
		u.RoleID = &adminRole.ID
	}

	u.IsApproved = false // Auto-approve all users
	u.Active = true

	err = h.uc.User.Client().Create(ctx.Request.Context(), u)
	if err != nil {
		h.l.Errorw("registration failed", "error", err)
		ctx.Redirect(http.StatusFound, "/admin/register?error=Registration+Failed:+"+err.Error())
		return
	}

	if isFirstUser {
		ctx.Redirect(http.StatusFound, "/admin/login?success=Admin+account+created+successfully")
	} else {
		ctx.Redirect(http.StatusFound, "/admin/login?success=Registration+successful.+You+can+now+login.")
	}
}

// Approvals (GET) - List users pending approval
func (h *Handler) Approvals(ctx *gin.Context) {
	pagination := h.bindPagination(ctx)

	// specific filter for unapproved users
	approved := false
	filter := &domain.UsersFilter{
		Pagination: pagination,
		UserFilter: domain.UserFilter{IsApproved: &approved},
	}

	users, count, err := h.uc.User.Client().Gets(ctx.Request.Context(), filter)
	if err != nil {
		h.l.Errorw("failed to fetch pending users", "error", err)
	}
	pagination.Total = int64(count)

	h.servePage(ctx, "approvals.html", "Approvals", "approvals", map[string]any{
		"Users":       users,
		"Pagination":  pagination,
		"QueryParams": ctx.Request.URL.Query(),
		"CSRF":        ctx.GetString("csrf"),
	})
}

// ApproveUser (POST) - Activate user
func (h *Handler) ApproveUser(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.Redirect(http.StatusFound, "/admin/approvals?error=Invalid+ID")
		return
	}

	err := h.uc.User.Client().ActivateUser(ctx.Request.Context(), id)
	if err != nil {
		h.l.Errorw("approval failed", "id", id, "error", err)
		ctx.Redirect(http.StatusFound, "/admin/approvals?error=Approval+Failed")
		return
	}

	ctx.Redirect(http.StatusFound, "/admin/approvals?success=User+Approved")
}

// CreateUser (GET) - Show Create User Form
func (h *Handler) CreateUser(ctx *gin.Context) {
	h.servePage(ctx, "users/create.html", "Create User", "users", map[string]any{
		"CSRF":  ctx.GetString("csrf"),
		"Error": ctx.Query("error"),
	})
}

// CreateUserPost (POST) - Create Active User
func (h *Handler) CreateUserPost(ctx *gin.Context) {
	email := ctx.PostForm("email")
	password := ctx.PostForm("password")
	phone := ctx.PostForm("phone")
	confirm := ctx.PostForm("confirm_password")

	if password != confirm {
		ctx.Redirect(http.StatusFound, "/admin/users/create?error=Passwords+do+not+match")
		return
	}

	u := domain.NewUser()
	u.Email = &email
	u.Phone = &phone
	u.Active = true
	u.IsApproved = true // Auto-approve since admin is creating it

	if err := u.SetPassword(password); err != nil {
		h.l.Errorw("failed to set password", "error", err)
		ctx.Redirect(http.StatusFound, "/admin/users/create?error=Invalid+Password")
		return
	}

	err := h.uc.User.Client().Create(ctx.Request.Context(), u)
	if err != nil {
		h.l.Errorw("admin user create failed", "error", err)
		ctx.Redirect(http.StatusFound, "/admin/users/create?error=Creation+Failed:+"+err.Error())
		return
	}

	ctx.Redirect(http.StatusFound, "/admin/users?success=User+Created")
}

func (h *Handler) Profile(ctx *gin.Context) {
	h.servePage(ctx, "profile.html", "Profile", "profile", nil)
}

func (h *Handler) UserAction(ctx *gin.Context) {
	idStr := ctx.Param("id")
	action := ctx.Param("action")

	id, err := uuid.Parse(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid user id"})
		return
	}

	var active bool
	switch action {
	case "block":
		active = false
	case "unblock":
		active = true
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid action"})
		return
	}

	err = h.uc.User.Client().SetStatus(ctx.Request.Context(), id, active)
	if err != nil {
		h.l.Errorw("user action failed", "id", id, "action", action, "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "success"})
}

func (h *Handler) UserDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	uid, err := uuid.Parse(idStr)
	if err != nil {
		h.l.Errorw("UserDetail - invalid user id", "id", idStr)
		ctx.Redirect(http.StatusFound, "/admin/users")
		return
	}

	// 1. Get User
	var user *domain.User

	// Optimistic: Check context for current user
	if val, ok := ctx.Get(consts.CtxUser); ok {
		if u, ok := val.(*domain.User); ok && u.ID == uid {
			user = u
		}
	}

	if user == nil {
		var err error
		user, err = h.uc.User.Client().Get(ctx.Request.Context(), &domain.UserFilter{ID: &uid})
		if err != nil {
			h.l.Errorw("UserDetail - failed to fetch user", "error", err)
			if apperrors.Is(err, apperrors.ErrServiceNotFound) {
				ctx.Redirect(http.StatusFound, "/admin/users")
				return
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to load user details",
				"details": err.Error(),
			})
			return
		}
	}

	// 2. Get Sessions
	sessions, _, err := h.uc.User.Session().Gets(ctx.Request.Context(), &domain.SessionsFilter{
		SessionFilter: domain.SessionFilter{UserID: &uid},
		Pagination:    &domain.Pagination{Limit: 20, Offset: 0},
	})
	if err != nil {
		h.l.Warnw("UserDetail - failed to fetch sessions", "error", err)
	}

	// 3. Get Role
	var role *domain.Role
	if user.RoleID != nil {
		r, err := h.uc.Authz.Role().Get(ctx.Request.Context(), &domain.RoleFilter{ID: user.RoleID})
		if err == nil {
			role = r
		}
	}

	h.servePage(ctx, "users/detail.html", "User Details", "users", map[string]any{
		"User":     user,
		"Sessions": sessions,
		"Role":     role,
	})
}

func (h *Handler) EditUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	uid, err := uuid.Parse(idStr)
	if err != nil {
		h.l.Errorw("EditUser - invalid user id", "id", idStr)
		ctx.Redirect(http.StatusFound, "/admin/users")
		return
	}

	user, err := h.uc.User.Client().Get(ctx.Request.Context(), &domain.UserFilter{ID: &uid})
	if err != nil {
		h.l.Errorw("EditUser - failed to fetch user", "error", err)
		ctx.Redirect(http.StatusFound, "/admin/users")
		return
	}

	h.servePage(ctx, "users/edit.html", "Edit User", "users", map[string]any{
		"User":  user,
		"CSRF":  ctx.GetString("csrf"),
		"Error": ctx.Query("error"),
	})
}

func (h *Handler) UpdateUserPost(ctx *gin.Context) {
	idStr := ctx.Param("id")
	uid, err := uuid.Parse(idStr)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/admin/users")
		return
	}

	// Fetch existing user
	user, err := h.uc.User.Client().Get(ctx.Request.Context(), &domain.UserFilter{ID: &uid})
	if err != nil {
		ctx.Redirect(http.StatusFound, "/admin/users")
		return
	}

	email := ctx.PostForm("email")
	phone := ctx.PostForm("phone")

	user.Email = &email
	user.Phone = &phone

	err = h.uc.User.Client().Update(ctx.Request.Context(), user)
	if err != nil {
		h.l.Errorw("UpdateUserPost - failed", "error", err)
		ctx.Redirect(http.StatusFound, "/admin/users/"+idStr+"/edit?error=Update+Failed")
		return
	}

	ctx.Redirect(http.StatusFound, "/admin/users/"+idStr+"?success=User+Updated")
}

func (h *Handler) BulkUsersAction(ctx *gin.Context) {
	var req struct {
		IDs    []string `json:"ids"`
		Action string   `json:"action"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request"})
		return
	}

	ctxCtx := ctx.Request.Context()
	for _, idStr := range req.IDs {
		uid, err := uuid.Parse(idStr)
		if err != nil {
			continue
		}

		switch req.Action {
		case "delete":
			if err := h.uc.User.Client().Delete(ctxCtx, &domain.UserFilter{ID: &uid}); err != nil {
				h.l.Warnw("failed to delete user in bulk", "user_id", uid, "error", err)
			}
		case "activate":
			if err := h.uc.User.Client().SetStatus(ctxCtx, uid, true); err != nil {
				h.l.Warnw("failed to activate user in bulk", "user_id", uid, "error", err)
			}
		case "deactivate":
			if err := h.uc.User.Client().SetStatus(ctxCtx, uid, false); err != nil {
				h.l.Warnw("failed to deactivate user in bulk", "user_id", uid, "error", err)
			}
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "success"})
}
