package restapi

import (
	"gct/config"
	"gct/internal/controller/restapi/v1/announcement"
	"gct/internal/controller/restapi/v1/audit"
	"gct/internal/controller/restapi/v1/authz"
	"gct/internal/controller/restapi/v1/dashboard"
	"gct/internal/controller/restapi/v1/dataexport"
	"gct/internal/controller/restapi/v1/emailtemplate"
	"gct/internal/controller/restapi/v1/errorcode"
	featureflagcrud "gct/internal/controller/restapi/v1/featureflagcrud"
	"gct/internal/controller/restapi/v1/integration"
	"gct/internal/controller/restapi/v1/iprule"
	"gct/internal/controller/restapi/v1/job"
	"gct/internal/controller/restapi/v1/minio"
	"gct/internal/controller/restapi/v1/notification"
	"gct/internal/controller/restapi/v1/ratelimit"
	"gct/internal/controller/restapi/v1/sitesetting"
	"gct/internal/controller/restapi/v1/test"
	"gct/internal/controller/restapi/v1/translation"
	"gct/internal/controller/restapi/v1/user"
	"gct/internal/controller/restapi/v1/webhook"
	"gct/internal/usecase"
	"gct/internal/shared/infrastructure/logger"
)

type Controller struct {
	User          *user.Controller
	Minio         *minio.Controller
	Authz         *authz.Controller
	Audit         *audit.Controller
	DataExport    dataexport.ControllerI
	EmailTemplate emailtemplate.ControllerI
	ErrorCode     *errorcode.Controller
	Integration   integration.ControllerI
	Translation   translation.ControllerI
	SiteSetting   sitesetting.ControllerI
	Dashboard     dashboard.ControllerI
	FeatureFlag   featureflagcrud.ControllerI
	RateLimit     ratelimit.ControllerI
	IPRule        iprule.ControllerI
	Webhook       webhook.ControllerI
	Job           job.ControllerI
	Announcement  announcement.ControllerI
	Notification  notification.ControllerI
	Test          *test.Controller
}

func NewController(uc *usecase.UseCase, cfg *config.Config, l logger.Log) *Controller {
	return &Controller{
		User:          user.New(uc, cfg, l),
		Minio:         minio.New(uc, l),
		Authz:         authz.New(uc, cfg, l),
		Audit:         audit.New(uc, cfg, l),
		DataExport:    dataexport.New(uc.DataExport, cfg, l),
		EmailTemplate: emailtemplate.New(uc.EmailTemplate, cfg, l),
		ErrorCode:     errorcode.New(uc.ErrorCode, l),
		Integration:   integration.New(uc.Integration, cfg, l),
		Translation:   translation.New(uc.Translation, cfg, l),
		SiteSetting:   sitesetting.New(uc.SiteSetting, cfg, l),
		Dashboard:     dashboard.New(uc.Dashboard, l),
		FeatureFlag:   featureflagcrud.New(uc.FeatureFlag, cfg, l),
		RateLimit:     ratelimit.New(uc.RateLimit, cfg, l),
		IPRule:        iprule.New(uc.IPRule, cfg, l),
		Webhook:       webhook.New(uc.Webhook, cfg, l),
		Job:           job.New(uc.Job, cfg, l),
		Announcement:  announcement.New(uc.Announcement, cfg, l),
		Notification:  notification.New(uc.Notification, cfg, l),
		Test:          test.New(uc, cfg, l),
	}
}
