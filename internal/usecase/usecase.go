package usecase

import (
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
)

// UseCase -.
type UseCase struct {
	Repo         *repo.Repo
	User         user.UseCaseI
	Minio        minio.Interface
	Authz        authz.UseCaseI
	Audit        audit.UseCaseI
	SiteSetting  sitesetting.UseCaseI
	ErrorCode    errorcode.UseCaseI
	Database     database.UseCaseI
	Integration   integration.UseCaseI
	Translation   translation.UseCaseI
	DataExport    dataexport.UseCaseI
	Dashboard     dashboard.UseCaseI
	EmailTemplate emailtemplate.UseCaseI
	FeatureFlag   featureflag.UseCaseI
	RateLimit    ratelimit.UseCaseI
	IPRule       iprule.UseCaseI
	Webhook      webhook.UseCaseI
	Job          job.UseCaseI
	Announcement announcement.UseCaseI
	Notification notification.UseCaseI
	File         file.UseCaseI
	UserSetting  usersetting.UseCaseI
	AsynqClient  *asynq.Client
}
