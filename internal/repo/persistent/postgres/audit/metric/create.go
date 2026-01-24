package metric

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
)

func (r *Repo) Create(ctx context.Context, m *domain.FunctionMetric) error {
	sql, args, err := r.builder.
		Insert(tableName).
		Columns(
			"name",
			"latency_ms",
			"is_panic",
			"panic_error",
			"created_at",
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
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build insert SQL query")
	}

	_, err = r.pool.Exec(ctx, sql, args...)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}
