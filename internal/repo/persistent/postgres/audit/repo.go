package audit

import (
	"gct/internal/repo/persistent/postgres/audit/history"
	"gct/internal/repo/persistent/postgres/audit/log"
	"gct/internal/repo/persistent/postgres/audit/metric"
	"gct/internal/repo/persistent/postgres/audit/systemerror"
	"gct/internal/shared/infrastructure/db/postgres"
	"gct/internal/shared/infrastructure/logger"
)

type Audit struct {
	Log         *log.Repo
	History     *history.Repo
	Metric      *metric.Repo
	SystemError *systemerror.Repo
}

func New(pg *postgres.Postgres, l logger.Log) *Audit {
	return &Audit{
		Log:         log.New(pg, l),
		History:     history.New(pg, l),
		Metric:      metric.New(pg, l),
		SystemError: systemerror.New(pg, l),
	}
}
