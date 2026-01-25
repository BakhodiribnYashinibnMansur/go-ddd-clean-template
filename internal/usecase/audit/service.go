package audit

import (
	"gct/internal/repo/persistent"
	"gct/internal/usecase/audit/auditlog"
	"gct/internal/usecase/audit/endpointhistory"
	"gct/internal/usecase/audit/metric"
	"gct/internal/usecase/audit/systemerror"
	"gct/pkg/logger"
)

type UseCase struct {
	Log         auditlog.UseCaseI
	History     endpointhistory.UseCaseI
	Metric      metric.UseCaseI
	SystemError systemerror.UseCaseI
}

func New(r *persistent.Repo, logger logger.Log) *UseCase {
	return &UseCase{
		Log:         auditlog.New(r, logger),
		History:     endpointhistory.New(r, logger),
		Metric:      metric.New(r, logger),
		SystemError: systemerror.New(r, logger),
	}
}
