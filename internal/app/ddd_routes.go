package app

import (
	"gct/internal/shared/infrastructure/logger"

	announcementhttp "gct/internal/announcement/interfaces/http"
	audithttp "gct/internal/audit/interfaces/http"
	authzhttp "gct/internal/authz/interfaces/http"
	dashboardhttp "gct/internal/dashboard/interfaces/http"
	dataexporthttp "gct/internal/dataexport/interfaces/http"
	emailtemplatehttp "gct/internal/emailtemplate/interfaces/http"
	errorcodehttp "gct/internal/errorcode/interfaces/http"
	featureflaghttp "gct/internal/featureflag/interfaces/http"
	filehttp "gct/internal/file/interfaces/http"
	integrationhttp "gct/internal/integration/interfaces/http"
	iprulehttp "gct/internal/iprule/interfaces/http"
	jobhttp "gct/internal/job/interfaces/http"
	metrichttp "gct/internal/metric/interfaces/http"
	notificationhttp "gct/internal/notification/interfaces/http"
	ratelimithttp "gct/internal/ratelimit/interfaces/http"
	sessionhttp "gct/internal/session/interfaces/http"
	sitesettinghttp "gct/internal/sitesetting/interfaces/http"
	systemerrorhttp "gct/internal/systemerror/interfaces/http"
	translationhttp "gct/internal/translation/interfaces/http"
	userhttp "gct/internal/user/interfaces/http"
	usersettinghttp "gct/internal/usersetting/interfaces/http"
	webhookhttp "gct/internal/webhook/interfaces/http"

	"github.com/gin-gonic/gin"
)

// RegisterDDDRoutes registers HTTP routes for all DDD bounded contexts.
func RegisterDDDRoutes(router *gin.Engine, bcs *DDDBoundedContexts, l logger.Log) {
	api := router.Group("/api/v2")

	// Core BCs
	userhttp.NewHandler(bcs.User, l).RegisterRoutes(api)
	authzhttp.NewHandler(bcs.Authz, l).RegisterRoutes(api)
	sessionhttp.NewHandler(bcs.Session, l).RegisterRoutes(api)

	// Supporting BCs
	audithttp.NewHandler(bcs.Audit, l).RegisterRoutes(api)
	dashboardhttp.NewHandler(bcs.Dashboard, l).RegisterRoutes(api)
	systemerrorhttp.NewHandler(bcs.SystemError, l).RegisterRoutes(api)
	metrichttp.NewHandler(bcs.Metric, l).RegisterRoutes(api)
	featureflaghttp.NewHandler(bcs.FeatureFlag, l).RegisterRoutes(api)
	integrationhttp.NewHandler(bcs.Integration, l).RegisterRoutes(api)
	webhookhttp.NewHandler(bcs.Webhook, l).RegisterRoutes(api)
	notificationhttp.NewHandler(bcs.Notification, l).RegisterRoutes(api)
	emailtemplatehttp.NewHandler(bcs.EmailTemplate, l).RegisterRoutes(api)
	announcementhttp.NewHandler(bcs.Announcement, l).RegisterRoutes(api)
	translationhttp.NewHandler(bcs.Translation, l).RegisterRoutes(api)
	sitesettinghttp.NewHandler(bcs.SiteSetting, l).RegisterRoutes(api)
	ratelimithttp.NewHandler(bcs.RateLimit, l).RegisterRoutes(api)
	iprulehttp.NewHandler(bcs.IPRule, l).RegisterRoutes(api)
	jobhttp.NewHandler(bcs.Job, l).RegisterRoutes(api)
	dataexporthttp.NewHandler(bcs.DataExport, l).RegisterRoutes(api)
	filehttp.NewHandler(bcs.File, l).RegisterRoutes(api)
	usersettinghttp.NewHandler(bcs.UserSetting, l).RegisterRoutes(api)
	errorcodehttp.NewHandler(bcs.ErrorCode, l).RegisterRoutes(api)
}
