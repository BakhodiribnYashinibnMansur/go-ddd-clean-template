package admin

import (
	"html/template"
	"math"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"gct/config"
	"gct/consts"
	"gct/internal/controller/restapi/middleware"
	"gct/internal/domain"
	"gct/internal/usecase"
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

func (h *Handler) Register(r *gin.RouterGroup, authmw *middleware.AuthMiddleware) {
	g := r.Group("/admin")
	g.GET("/login", h.Login)

	protected := g.Group("/")
	protected.Use(authmw.AuthWeb)

	protected.GET("/dashboard", h.Dashboard)
	protected.GET("/users", h.Users)
	protected.GET("/sessions", h.Sessions)
	protected.GET("/rbac/roles", h.Roles)
	protected.GET("/rbac/permissions", h.Permissions)
	protected.GET("/rbac/scopes", h.Scopes)
	protected.GET("/abac/policies", h.Policies)
	protected.GET("/system-errors", h.SystemErrors)
	protected.GET("/functions", h.FunctionMetrics)
	protected.GET("/audit/logs", h.AuditLogs)
	protected.GET("/audit/history", h.EndpointHistory)
	protected.GET("/linter", h.Linter)
	// Asynq Monitor
	mon := h.NewAsynqMonitor()
	protected.Any("/asynq/*filepath", gin.WrapH(mon))
	protected.GET("/asynq", gin.WrapH(mon))

	protected.GET("/settings", h.Settings)
}

func (h *Handler) Login(ctx *gin.Context) {
	tmpl, err := template.ParseFiles("internal/web/admin/templates/pages/login.html")
	if err != nil {
		h.l.Errorw("failed to parse login template", "error", err)
		ctx.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}
	// TODO: Inject CSRF token
	if err := tmpl.Execute(ctx.Writer, map[string]any{"Error": ctx.Query("error")}); err != nil {
		h.l.Errorw("failed to execute login template", "error", err)
	}
}

// PageData holds common data for all admin pages
type PageData struct {
	Title          string
	ActivePage     string
	User           *domain.User
	Can            map[string]bool
	TracingEnabled bool
	JaegerURL      string
	Data           any
}

func (h *Handler) render(ctx *gin.Context, tmplName string, data PageData) {
	files := []string{
		"internal/web/admin/templates/layout/base.html",
		"internal/web/admin/templates/layout/header.html",
		"internal/web/admin/templates/layout/sidebar.html",
		"internal/web/admin/templates/layout/pagination.html",
		filepath.Join("internal/web/admin/templates/pages", tmplName),
	}

	tmpl, err := template.New("base").Funcs(template.FuncMap{
		"currUserEmail": func() string {
			if data.User != nil && data.User.Email != nil {
				return *data.User.Email
			}
			return "Admin"
		},
		"add": func(a, b int64) int64 { return a + b },
		"sub": func(a, b int64) int64 { return a - b },
		"seq": func(start, end int64) []int64 {
			var res []int64
			for i := start; i <= end; i++ {
				res = append(res, i)
			}
			return res
		},
		"totalPages": func(total, limit int64) int64 {
			if limit == 0 {
				return 1
			}
			return int64(math.Ceil(float64(total) / float64(limit)))
		},
		"currPage": func(offset, limit int64) int64 {
			if limit == 0 {
				return 1
			}
			return (offset / limit) + 1
		},
		"paginationLink": func(currURL url.Values, page int64, limit int64) string {
			currURL.Set("page", strconv.FormatInt(page, 10))
			currURL.Set("limit", strconv.FormatInt(limit, 10))
			return "?" + currURL.Encode()
		},
	}).ParseFiles(files...)
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
	data := map[string]any{
		"CurrentDate": time.Now().Format("Monday, January 2, 2006"),
	}
	h.servePage(ctx, "dashboard.html", "Dashboard", "dashboard", data)
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

	users, count, err := h.uc.User.Client.Gets(ctx.Request.Context(), filter)
	if err != nil {
		h.l.Errorw("failed to fetch users", "error", err)
	}
	pagination.Total = int64(count)

	h.servePage(ctx, "users.html", "Users", "users", map[string]any{
		"Users":       users,
		"Pagination":  pagination,
		"Filter":      filter,
		"QueryParams": ctx.Request.URL.Query(),
	})
}

func (h *Handler) Sessions(ctx *gin.Context) {
	pagination := h.bindPagination(ctx)
	filter := &domain.SessionsFilter{Pagination: pagination}

	if uid := ctx.Query("user_id"); uid != "" {
		if id, err := uuid.Parse(uid); err == nil {
			filter.UserID = &id
		}
	}

	sessions, count, err := h.uc.User.Session.Gets(ctx.Request.Context(), filter)
	if err != nil {
		h.l.Errorw("failed to fetch sessions", "error", err)
	}
	pagination.Total = int64(count)

	h.servePage(ctx, "sessions.html", "Sessions", "sessions", map[string]any{
		"Sessions":    sessions,
		"Pagination":  pagination,
		"QueryParams": ctx.Request.URL.Query(),
	})
}

func (h *Handler) Roles(ctx *gin.Context) {
	pagination := h.bindPagination(ctx)
	roles, count, err := h.uc.Authz.Role.Gets(ctx.Request.Context(), &domain.RolesFilter{
		Pagination: pagination,
	})
	if err != nil {
		h.l.Errorw("failed to fetch roles", "error", err)
	}
	pagination.Total = int64(count)

	h.servePage(ctx, "rbac/roles.html", "Roles", "roles", map[string]any{
		"Roles":       roles,
		"Pagination":  pagination,
		"QueryParams": ctx.Request.URL.Query(),
	})
}

func (h *Handler) Permissions(ctx *gin.Context) {
	pagination := h.bindPagination(ctx)
	perms, count, err := h.uc.Authz.Permission.Gets(ctx.Request.Context(), &domain.PermissionsFilter{
		Pagination: pagination,
	})
	if err != nil {
		h.l.Errorw("failed to fetch permissions", "error", err)
	}
	pagination.Total = int64(count)

	h.servePage(ctx, "rbac/permissions.html", "Permissions", "permissions", map[string]any{
		"Permissions": perms,
		"Pagination":  pagination,
		"QueryParams": ctx.Request.URL.Query(),
	})
}

func (h *Handler) Scopes(ctx *gin.Context) {
	pagination := h.bindPagination(ctx)
	scopes, count, err := h.uc.Authz.Scope.Gets(ctx.Request.Context(), &domain.ScopesFilter{
		Pagination: pagination,
	})
	if err != nil {
		h.l.Errorw("failed to fetch scopes", "error", err)
	}
	pagination.Total = int64(count)

	h.servePage(ctx, "rbac/scopes.html", "Scopes", "scopes", map[string]any{
		"Scopes":      scopes,
		"Pagination":  pagination,
		"QueryParams": ctx.Request.URL.Query(),
	})
}

func (h *Handler) Policies(ctx *gin.Context) {
	pagination := h.bindPagination(ctx)
	policies, count, err := h.uc.Authz.Policy.Gets(ctx.Request.Context(), &domain.PoliciesFilter{
		Pagination: pagination,
	})
	if err != nil {
		h.l.Errorw("failed to fetch policies", "error", err)
	}
	pagination.Total = int64(count)

	h.servePage(ctx, "abac/policies.html", "Policies", "policies", map[string]any{
		"Policies":    policies,
		"Pagination":  pagination,
		"QueryParams": ctx.Request.URL.Query(),
	})
}

func (h *Handler) Linter(ctx *gin.Context) {
	h.servePage(ctx, "linter.html", "Code Linter", "linter", nil)
}

func (h *Handler) Settings(ctx *gin.Context) {
	h.servePage(ctx, "settings.html", "Settings", "settings", nil)
}

func (h *Handler) servePage(ctx *gin.Context, tmplFile, title, activePage string, pageContent any) {
	sessionVal, _ := ctx.Get(consts.CtxSession)
	sessObj, ok := sessionVal.(*domain.Session)
	if !ok {
		h.l.Warnw("invalid session object in context")
		ctx.String(http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.uc.User.Client.Get(ctx, &domain.UserFilter{ID: &sessObj.UserID})
	if err != nil {
		h.l.Warnw("failed to fetch current user", "user_id", sessObj.UserID)
	}

	path := ctx.Request.URL.Path
	method := ctx.Request.Method

	env := map[string]any{
		consts.PolicyKeyIP:        ctx.ClientIP(),
		consts.PolicyKeyUserAgent: ctx.GetHeader("User-Agent"),
		consts.PolicyKeyTime:      time.Now(),
		consts.PolicyKeyUserID:    sessObj.UserID,
	}
	if user != nil && user.RoleID != nil {
		env[consts.PolicyKeyRoleID] = *user.RoleID
	}

	allowed, err := h.uc.Authz.Access.Check(ctx.Request.Context(), sessObj.UserID, sessObj, path, method, env)
	if err != nil {
		h.l.Errorw("access check error", "error", err)
		ctx.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if !allowed {
		ctx.String(http.StatusForbidden, "Access Denied")
		return
	}

	can := make(map[string]bool)
	checks := map[string]string{
		"users_view":            "/admin/users",
		"sessions_view":         "/admin/sessions",
		"roles_view":            "/admin/rbac/roles",
		"permissions_view":      "/admin/rbac/permissions",
		"scopes_view":           "/admin/rbac/scopes",
		"policies_view":         "/admin/abac/policies",
		"system_errors_view":    "/admin/system-errors",
		"functions_view":        "/admin/functions",
		"audit_logs_view":       "/admin/audit/logs",
		"endpoint_history_view": "/admin/audit/history",
		"linter_view":           "/admin/linter",
		"asynq_view":            "/admin/asynq",
		"settings_view":         "/admin/settings",
	}

	for key, checkPath := range checks {
		ok, _ := h.uc.Authz.Access.Check(ctx.Request.Context(), sessObj.UserID, sessObj, checkPath, "GET", env)
		can[key] = ok
	}

	h.render(ctx, tmplFile, PageData{
		Title:          title,
		ActivePage:     activePage,
		User:           user,
		Can:            can,
		TracingEnabled: h.cfg.Tracing.Enabled,
		JaegerURL:      h.cfg.Tracing.HttpEndpoint,
		Data:           pageContent,
	})
}
