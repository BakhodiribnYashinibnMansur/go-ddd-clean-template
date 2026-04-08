package app

import (
	"gct/internal/kernel/infrastructure/logger"

	miniogo "github.com/minio/minio-go/v7"

	activityloghttp "gct/internal/context/ops/supporting/activitylog/interfaces/http"
	announcementhttp "gct/internal/context/content/supporting/announcement/interfaces/http"
	audithttp "gct/internal/context/iam/supporting/audit/interfaces/http"
	authzhttp "gct/internal/context/iam/generic/authz/interfaces/http"
	dataexporthttp "gct/internal/context/admin/supporting/dataexport/interfaces/http"
	statisticshttp "gct/internal/context/admin/supporting/statistics/interfaces/http"
	errorcodehttp "gct/internal/context/admin/supporting/errorcode/interfaces/http"
	featureflaghttp "gct/internal/context/admin/generic/featureflag/interfaces/http"
	filehttp "gct/internal/context/content/generic/file/interfaces/http"
	integrationhttp "gct/internal/context/admin/supporting/integration/interfaces/http"
	iprulehttp "gct/internal/context/ops/supporting/iprule/interfaces/http"

	metrichttp "gct/internal/context/ops/generic/metric/interfaces/http"
	notificationhttp "gct/internal/context/content/generic/notification/interfaces/http"
	ratelimithttp "gct/internal/context/ops/generic/ratelimit/interfaces/http"
	sessionhttp "gct/internal/context/iam/generic/session/interfaces/http"
	sitesettinghttp "gct/internal/context/admin/supporting/sitesetting/interfaces/http"
	systemerrorhttp "gct/internal/context/ops/generic/systemerror/interfaces/http"
	translationhttp "gct/internal/context/content/generic/translation/interfaces/http"
	userhttp "gct/internal/context/iam/generic/user/interfaces/http"
	usersettinghttp "gct/internal/context/iam/generic/usersetting/interfaces/http"

	"github.com/gin-gonic/gin"
)

// RouteOptions holds optional dependencies for route registration.
type RouteOptions struct {
	Minio       *miniogo.Client
	MinioBucket string
}

// dddHandlers bundles every HTTP handler constructed from the DDD bounded contexts.
type dddHandlers struct {
	user         *userhttp.Handler
	authz        *authzhttp.Handler
	session      *sessionhttp.Handler
	audit        *audithttp.Handler
	statistics   *statisticshttp.Handler
	sysErr       *systemerrorhttp.Handler
	metric       *metrichttp.Handler
	ff           *featureflaghttp.Handler
	integration  *integrationhttp.Handler
	notification *notificationhttp.Handler
	announcement *announcementhttp.Handler
	translation  *translationhttp.Handler
	siteSetting  *sitesettinghttp.Handler
	rateLimit    *ratelimithttp.Handler
	ipRule       *iprulehttp.Handler
	dataExport   *dataexporthttp.Handler
	file         *filehttp.Handler
	userSetting  *usersettinghttp.Handler
	errorCode    *errorcodehttp.Handler
	activityLog  *activityloghttp.Handler
}

// buildDDDHandlers constructs every HTTP handler for the DDD bounded contexts.
func buildDDDHandlers(bcs *DDDBoundedContexts, l logger.Log, opt RouteOptions) *dddHandlers {
	fileHandler := filehttp.NewHandler(bcs.File, l)
	if opt.Minio != nil {
		fileHandler.SetMinio(opt.Minio, opt.MinioBucket)
	}
	return &dddHandlers{
		user:         userhttp.NewHandler(bcs.User, l),
		authz:        authzhttp.NewHandler(bcs.Authz, l),
		session:      sessionhttp.NewHandler(bcs.Session, l),
		audit:        audithttp.NewHandler(bcs.Audit, l),
		statistics:   statisticshttp.NewHandler(bcs.Statistics, l),
		sysErr:       systemerrorhttp.NewHandler(bcs.SystemError, l),
		metric:       metrichttp.NewHandler(bcs.Metric, l),
		ff:           featureflaghttp.NewHandler(bcs.FeatureFlag, l),
		integration:  integrationhttp.NewHandler(bcs.Integration, l),
		notification: notificationhttp.NewHandler(bcs.Notification, l),
		announcement: announcementhttp.NewHandler(bcs.Announcement, l),
		translation:  translationhttp.NewHandler(bcs.Translation, l),
		siteSetting:  sitesettinghttp.NewHandler(bcs.SiteSetting, l),
		rateLimit:    ratelimithttp.NewHandler(bcs.RateLimit, l),
		ipRule:       iprulehttp.NewHandler(bcs.IPRule, l),
		dataExport:   dataexporthttp.NewHandler(bcs.DataExport, l),
		file:         fileHandler,
		userSetting:  usersettinghttp.NewHandler(bcs.UserSetting, l),
		errorCode:    errorcodehttp.NewHandler(bcs.ErrorCode, l),
		activityLog:  activityloghttp.NewHandler(bcs.ActivityLog, l),
	}
}

// RegisterDDDRoutes registers all HTTP routes for DDD bounded contexts.
func RegisterDDDRoutes(
	router *gin.Engine,
	bcs *DDDBoundedContexts,
	authMW, authzMW, csrfMW gin.HandlerFunc,
	l logger.Log,
	opts ...RouteOptions,
) {
	var opt RouteOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	h := buildDDDHandlers(bcs, l, opt)

	registerAuthRoutes(router, authMW, h.user)

	protected := router.Group("/api/v1")
	protected.Use(authMW, authzMW, csrfMW)

	registerIAMRoutes(protected, h)
	registerAuthzRoutes(protected, h)
	registerOpsRoutes(protected, h)
	registerContentRoutes(protected, h)
	registerAdminRoutes(protected, h)
	h.ff.RegisterRoutes(protected)
}

// registerAuthRoutes wires the public and auth-only sign-in/sign-up/sign-out endpoints.
func registerAuthRoutes(router *gin.Engine, authMW gin.HandlerFunc, userHandler *userhttp.Handler) {
	public := router.Group("/api/v1/auth")
	{
		public.POST("/sign-in", userHandler.SignIn)
		public.POST("/sign-up", userHandler.SignUp)
	}

	authOnly := router.Group("/api/v1/auth")
	authOnly.Use(authMW)
	{
		authOnly.POST("/sign-out", userHandler.SignOut)
	}
}

// registerIAMRoutes wires identity routes (users, sessions, audit logs).
func registerIAMRoutes(protected *gin.RouterGroup, h *dddHandlers) {
	users := protected.Group("/users")
	{
		users.POST("", h.user.Create)
		users.GET("", h.user.List)
		users.GET("/:id", h.user.Get)
		users.PATCH("/:id", h.user.Update)
		users.DELETE("/:id", h.user.Delete)
		users.POST("/:id/approve", h.user.Approve)
		users.POST("/:id/role", h.user.ChangeRole)
		users.POST("/bulk-action", h.user.BulkAction)
	}

	sessions := protected.Group("/sessions")
	{
		sessions.GET("", h.session.List)
		sessions.GET("/:id", h.session.Get)
		sessions.DELETE("/:id", h.session.Delete)
		sessions.POST("/revoke-all", h.session.RevokeAll)
	}

	auditLogs := protected.Group("/audit-logs")
	{
		auditLogs.GET("", h.audit.ListAuditLogs)
	}

	endpointHistory := protected.Group("/endpoint-history")
	{
		endpointHistory.GET("", h.audit.ListEndpointHistory)
	}
}

// registerAuthzRoutes wires authorization routes (roles, permissions, policies, scopes).
func registerAuthzRoutes(protected *gin.RouterGroup, h *dddHandlers) {
	roles := protected.Group("/roles")
	{
		roles.POST("", h.authz.CreateRole)
		roles.GET("", h.authz.ListRoles)
		roles.GET("/:id", h.authz.GetRole)
		roles.PATCH("/:id", h.authz.UpdateRole)
		roles.DELETE("/:id", h.authz.DeleteRole)
		roles.POST("/:id/permissions", h.authz.AssignPermission)
	}

	permissions := protected.Group("/permissions")
	{
		permissions.POST("", h.authz.CreatePermission)
		permissions.GET("", h.authz.ListPermissions)
		permissions.DELETE("/:id", h.authz.DeletePermission)
		permissions.POST("/:id/scopes", h.authz.AssignScope)
	}

	policies := protected.Group("/policies")
	{
		policies.POST("", h.authz.CreatePolicy)
		policies.GET("", h.authz.ListPolicies)
		policies.PATCH("/:id", h.authz.UpdatePolicy)
		policies.DELETE("/:id", h.authz.DeletePolicy)
		policies.POST("/:id/toggle", h.authz.TogglePolicy)
	}

	scopes := protected.Group("/scopes")
	{
		scopes.POST("", h.authz.CreateScope)
		scopes.GET("", h.authz.ListScopes)
		scopes.DELETE("", h.authz.DeleteScope)
	}
}

// registerOpsRoutes wires operations routes (system errors, metrics, rate limits, ip rules).
func registerOpsRoutes(protected *gin.RouterGroup, h *dddHandlers) {
	systemErrors := protected.Group("/system-errors")
	{
		systemErrors.POST("", h.sysErr.Create)
		systemErrors.GET("", h.sysErr.List)
		systemErrors.GET("/:id", h.sysErr.Get)
		systemErrors.POST("/:id/resolve", h.sysErr.Resolve)
	}

	metrics := protected.Group("/metrics")
	{
		metrics.POST("", h.metric.Create)
		metrics.GET("", h.metric.List)
	}

	rateLimits := protected.Group("/rate-limits")
	{
		rateLimits.POST("", h.rateLimit.Create)
		rateLimits.GET("", h.rateLimit.List)
		rateLimits.GET("/:id", h.rateLimit.Get)
		rateLimits.PATCH("/:id", h.rateLimit.Update)
		rateLimits.DELETE("/:id", h.rateLimit.Delete)
	}

	ipRules := protected.Group("/ip-rules")
	{
		ipRules.POST("", h.ipRule.Create)
		ipRules.GET("", h.ipRule.List)
		ipRules.GET("/:id", h.ipRule.Get)
		ipRules.PATCH("/:id", h.ipRule.Update)
		ipRules.DELETE("/:id", h.ipRule.Delete)
	}

	h.activityLog.RegisterRoutes(protected)
}

// registerContentRoutes wires content routes (notifications, announcements, translations, files).
func registerContentRoutes(protected *gin.RouterGroup, h *dddHandlers) {
	notifications := protected.Group("/notifications")
	{
		notifications.POST("", h.notification.Create)
		notifications.GET("", h.notification.List)
		notifications.GET("/:id", h.notification.Get)
		notifications.DELETE("/:id", h.notification.Delete)
	}

	announcements := protected.Group("/announcements")
	{
		announcements.POST("", h.announcement.Create)
		announcements.GET("", h.announcement.List)
		announcements.GET("/:id", h.announcement.Get)
		announcements.PATCH("/:id", h.announcement.Update)
		announcements.DELETE("/:id", h.announcement.Delete)
	}

	translations := protected.Group("/translations")
	{
		translations.POST("", h.translation.Create)
		translations.GET("", h.translation.List)
		translations.GET("/:id", h.translation.Get)
		translations.PATCH("/:id", h.translation.Update)
		translations.DELETE("/:id", h.translation.Delete)
	}

	files := protected.Group("/files")
	{
		files.POST("", h.file.Create)
		files.GET("", h.file.List)
		files.GET("/:id", h.file.Get)
		files.POST("/upload/image", h.file.UploadImage)
		files.POST("/upload/images", h.file.UploadImages)
		files.POST("/upload/doc", h.file.UploadDoc)
		files.GET("/download", h.file.Download)
	}
}

// registerAdminRoutes wires admin routes (statistics, integrations, site settings, data exports,
// user settings, error codes).
func registerAdminRoutes(protected *gin.RouterGroup, h *dddHandlers) {
	h.statistics.RegisterRoutes(protected)

	integrations := protected.Group("/integrations")
	{
		integrations.POST("", h.integration.Create)
		integrations.GET("", h.integration.List)
		integrations.GET("/:id", h.integration.Get)
		integrations.PATCH("/:id", h.integration.Update)
		integrations.DELETE("/:id", h.integration.Delete)
	}

	siteSettings := protected.Group("/site-settings")
	{
		siteSettings.POST("", h.siteSetting.Create)
		siteSettings.GET("", h.siteSetting.List)
		siteSettings.GET("/:id", h.siteSetting.Get)
		siteSettings.PATCH("/:id", h.siteSetting.Update)
		siteSettings.DELETE("/:id", h.siteSetting.Delete)
	}

	dataExports := protected.Group("/data-exports")
	{
		dataExports.POST("", h.dataExport.Create)
		dataExports.GET("", h.dataExport.List)
		dataExports.GET("/:id", h.dataExport.Get)
		dataExports.PATCH("/:id", h.dataExport.Update)
		dataExports.DELETE("/:id", h.dataExport.Delete)
	}

	userSettings := protected.Group("/user-settings")
	{
		userSettings.POST("", h.userSetting.Upsert)
		userSettings.GET("", h.userSetting.List)
		userSettings.DELETE("/:id", h.userSetting.Delete)
	}

	errorCodes := protected.Group("/error-codes")
	{
		errorCodes.POST("", h.errorCode.Create)
		errorCodes.GET("", h.errorCode.List)
		errorCodes.GET("/:id", h.errorCode.Get)
		errorCodes.PATCH("/:id", h.errorCode.Update)
		errorCodes.DELETE("/:id", h.errorCode.Delete)
	}
}
