package usecase

import (
	"gct/config"
	"gct/internal/repo"
	"gct/internal/usecase/announcement"
	"gct/internal/usecase/audit"
	"gct/internal/usecase/authz"
	"gct/internal/usecase/dashboard"
	"gct/internal/usecase/database"
	errorcode "gct/internal/usecase/errorcode"
	"gct/internal/usecase/dataexport"
	"gct/internal/usecase/emailtemplate"
	"gct/internal/usecase/featureflag"
	"gct/internal/usecase/integration"
	"gct/internal/usecase/iprule"
	"gct/internal/usecase/job"
	"gct/internal/usecase/file"
	"gct/internal/usecase/minio"
	"gct/internal/usecase/notification"
	"gct/internal/usecase/ratelimit"
	"gct/internal/usecase/sitesetting"
	"gct/internal/usecase/translation"
	"gct/internal/usecase/user"
	"gct/internal/usecase/usersetting"
	"gct/internal/usecase/webhook"
	"gct/internal/shared/infrastructure/asynq"
	"gct/internal/shared/infrastructure/logger"
)

// NewUseCase -.
func NewUseCase(repos *repo.Repo, logger logger.Log, cfg *config.Config, asynqClient *asynq.Client) *UseCase {
	pg := repos.Persistent.Postgres
	return &UseCase{
		Repo:         repos,
		User:         user.New(repos, logger, cfg),
		Minio:        minio.New(repos, logger),
		Authz:        authz.New(repos, logger, cfg),
		Audit:        audit.New(repos.Persistent, logger),
		SiteSetting:  sitesetting.New(repos.Persistent, logger),
		ErrorCode:    errorcode.New(repos, logger),
		Database:     database.New(pg, logger, cfg),
		Integration:   integration.New(pg.Integration, logger, cfg),
		Translation:   translation.New(pg.Translation, logger),
		DataExport:    dataexport.New(pg.DataExport, logger, cfg),
		EmailTemplate: emailtemplate.New(pg.EmailTemplate, logger, cfg),
		FeatureFlag:   featureflag.New(pg.FeatureFlag, logger, cfg),
		RateLimit:    ratelimit.New(pg.RateLimit, logger, cfg),
		IPRule:       iprule.New(pg.IPRule, logger, cfg),
		Webhook:      webhook.New(pg.Webhook, logger, cfg),
		Job:          job.New(pg.Job, logger, cfg),
		Announcement: announcement.New(pg.Announcement, logger, cfg),
		Notification: notification.New(pg.Notification, logger, cfg),
		Dashboard:    dashboard.New(pg.Dashboard, logger),
		File:         file.New(pg.FileMetadata, logger),
		UserSetting:  usersetting.New(repos.Persistent.Postgres.User.Setting, logger),
		AsynqClient:  asynqClient,
	}
}
