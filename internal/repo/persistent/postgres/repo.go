package postgres

import (
	"context"

	"gct/internal/repo/persistent/postgres/announcement"
	"gct/internal/repo/persistent/postgres/audit"
	"gct/internal/repo/persistent/postgres/authz"
	"gct/internal/repo/persistent/postgres/dashboard"
	"gct/internal/repo/persistent/postgres/dataexport"
	errorcode "gct/internal/repo/persistent/postgres/errorcode"
	"gct/internal/repo/persistent/postgres/emailtemplate"
	"gct/internal/repo/persistent/postgres/featureflag"
	"gct/internal/repo/persistent/postgres/filemetadata"
	"gct/internal/repo/persistent/postgres/integration"
	"gct/internal/repo/persistent/postgres/iprule"
	"gct/internal/repo/persistent/postgres/job"
	"gct/internal/repo/persistent/postgres/notification"
	"gct/internal/repo/persistent/postgres/ratelimit"
	sitesetting "gct/internal/repo/persistent/postgres/sitesetting"
	systemerror "gct/internal/repo/persistent/postgres/systemerror"
	"gct/internal/repo/persistent/postgres/translation"
	"gct/internal/repo/persistent/postgres/user"
	"gct/internal/repo/persistent/postgres/webhook"
	announcementUC "gct/internal/usecase/announcement"
	ucdashboard "gct/internal/usecase/dashboard"
	ucdataexport "gct/internal/usecase/dataexport"
	ucemailtemplate "gct/internal/usecase/emailtemplate"
	ucfeatureflag "gct/internal/usecase/featureflag"
	ucfile "gct/internal/usecase/file"
	integrationUC "gct/internal/usecase/integration"
	uciprule "gct/internal/usecase/iprule"
	ucjob "gct/internal/usecase/job"
	ucnotification "gct/internal/usecase/notification"
	ucratelimit "gct/internal/usecase/ratelimit"
	translationUC "gct/internal/usecase/translation"
	ucwebhook "gct/internal/usecase/webhook"
	"gct/internal/shared/infrastructure/db/postgres"
	"gct/internal/shared/infrastructure/logger"

	"github.com/Masterminds/squirrel"
)

type Repo struct {
	User         *user.User
	Authz        *authz.Authz
	Audit        *audit.Audit
	SiteSetting  *sitesetting.Repo
	SystemError  *systemerror.Repo
	ErrorCode    *errorcode.Repo
	Integration   integrationUC.Repository
	Translation   translationUC.Repository
	DataExport    ucdataexport.Repository
	EmailTemplate ucemailtemplate.Repository
	FeatureFlag   ucfeatureflag.Repository
	RateLimit     ucratelimit.Repository
	IPRule        uciprule.Repository
	Webhook       ucwebhook.Repository
	Job           ucjob.Repository
	Announcement  announcementUC.Repository
	Notification  ucnotification.Repository
	Dashboard     ucdashboard.Repository
	FileMetadata  ucfile.Repository
	DB            *postgres.Postgres
}

func New(pg *postgres.Postgres, logger logger.Log) (*Repo, error) {
	pg.Builder.PlaceholderFormat(squirrel.Dollar)
	return &Repo{
		User:         user.New(pg, logger),
		Authz:        authz.New(pg, logger),
		Audit:        audit.New(pg, logger),
		SiteSetting:  sitesetting.New(pg.Pool),
		SystemError:  systemerror.New(pg, logger),
		ErrorCode:    errorcode.New(pg, logger),
		Integration:   integration.New(pg, logger),
		Translation:   translation.New(pg, logger),
		DataExport:    dataexport.New(pg, logger),
		EmailTemplate: emailtemplate.New(pg, logger),
		FeatureFlag:   featureflag.New(pg, logger),
		RateLimit:     ratelimit.New(pg, logger),
		IPRule:        iprule.New(pg, logger),
		Webhook:       webhook.New(pg, logger),
		Job:           job.New(pg, logger),
		Announcement:  announcement.New(pg, logger),
		Notification:  notification.New(pg, logger),
		Dashboard:     dashboard.New(pg, logger),
		FileMetadata:  filemetadata.New(pg, logger),
		DB:            pg,
	}, nil
}

func (r *Repo) Ping(ctx context.Context) error {
	return r.DB.Pool.Ping(ctx)
}
