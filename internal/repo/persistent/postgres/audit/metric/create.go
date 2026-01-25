package metric

import (
	"context"

	"gct/consts"
	"gct/internal/domain"
	"gct/internal/repo/schema"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, m *domain.FunctionMetric) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns(
			schema.FunctionMetricName,
			schema.FunctionMetricLatencyMs,
			schema.FunctionMetricIsPanic,
			schema.FunctionMetricPanicError,
			schema.FunctionMetricCreatedAt,
		).
		Values(
			m.Name,
			m.LatencyMs,
			m.IsPanic,
			m.PanicError,
			m.CreatedAt,
		).
		ToSql()
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, consts.ErrMsgFailedToBuildInsert)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
