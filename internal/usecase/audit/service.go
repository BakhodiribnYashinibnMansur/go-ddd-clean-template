package audit

import (
	"gct/internal/repo/persistent"
	"gct/internal/usecase/audit/auditLog"
	"gct/internal/usecase/audit/endpointHistory"
	"gct/internal/usecase/audit/metric"
	"gct/internal/usecase/audit/systemError"
	"gct/pkg/logger"
)

type UseCase struct {
	Log         auditLog.UseCaseI
	History     endpointHistory.UseCaseI
	Metric      metric.UseCaseI
	SystemError systemError.UseCaseI
}

func New(r *persistent.Repo, logger logger.Log) *UseCase {
	return &UseCase{
		Log:         auditLog.New(r, logger),
		History:     endpointHistory.New(r, logger),
		Metric:      metric.New(r, logger),
		SystemError: systemError.New(r, logger),
	}
}
