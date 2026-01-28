package audit

import (
	"gct/internal/repo/persistent"
	"gct/internal/usecase/audit/auditlog"
	"gct/internal/usecase/audit/endpointhistory"
	"gct/internal/usecase/audit/metric"
	"gct/internal/usecase/audit/systemerror"
	"gct/pkg/logger"
)

type UseCaseI interface {
	Log() auditlog.UseCaseI
	History() endpointhistory.UseCaseI
	Metric() metric.UseCaseI
	SystemError() systemerror.UseCaseI
}

type UseCase struct {
	log         auditlog.UseCaseI
	history     endpointhistory.UseCaseI
	metric      metric.UseCaseI
	systemError systemerror.UseCaseI
}

func New(r *persistent.Repo, logger logger.Log) UseCaseI {
	return &UseCase{
		log:         auditlog.New(r, logger),
		history:     endpointhistory.New(r, logger),
		metric:      metric.New(r, logger),
		systemError: systemerror.New(r, logger),
	}
}

func (uc *UseCase) Log() auditlog.UseCaseI            { return uc.log }
func (uc *UseCase) History() endpointhistory.UseCaseI { return uc.history }
func (uc *UseCase) Metric() metric.UseCaseI           { return uc.metric }
func (uc *UseCase) SystemError() systemerror.UseCaseI { return uc.systemError }
