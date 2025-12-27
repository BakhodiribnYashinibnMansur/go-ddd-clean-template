package translation

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/evrone/go-clean-template/pkg/db/postgres"
	"go.uber.org/zap"
)

const _defaultEntityCap = 64

// TranslationRepo -.
type TranslationRepo struct {
	*postgres.Postgres
	logger *zap.Logger
}

// NewTranslationRepo -.
func NewTranslationRepo(pg *postgres.Postgres, logger *zap.Logger) *TranslationRepo {
	return &TranslationRepo{
		Postgres: pg,
		logger:   logger,
	}
}

// GetHistory -.
func (r *TranslationRepo) GetHistory(ctx context.Context) ([]domain.Translation, error) {
	r.logger.Info("TranslationRepo.GetHistory started")

	sql, _, err := r.Builder.
		Select("source, destination, original, translation").
		From("history").
		ToSql()
	if err != nil {
		r.logger.Error("TranslationRepo.GetHistory - r.Builder", zap.Error(err))
		return nil, fmt.Errorf("TranslationRepo - GetHistory - r.Builder: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql)
	if err != nil {
		r.logger.Error("TranslationRepo.GetHistory - r.Pool.Query", zap.Error(err))
		return nil, fmt.Errorf("TranslationRepo - GetHistory - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	entities := make([]domain.Translation, 0, _defaultEntityCap)

	for rows.Next() {
		e := domain.Translation{}

		err = rows.Scan(&e.Source, &e.Destination, &e.Original, &e.Translation)
		if err != nil {
			r.logger.Error("TranslationRepo.GetHistory - rows.Scan", zap.Error(err))
			return nil, fmt.Errorf("TranslationRepo - GetHistory - rows.Scan: %w", err)
		}

		entities = append(entities, e)
	}

	r.logger.Info("TranslationRepo.GetHistory finished", zap.Int("count", len(entities)))
	return entities, nil
}

// Store -.
func (r *TranslationRepo) Store(ctx context.Context, t domain.Translation) error {
	r.logger.Info("TranslationRepo.Store started",
		zap.String("source", t.Source),
		zap.String("destination", t.Destination),
		zap.String("original", t.Original))

	sql, args, err := r.Builder.
		Insert("history").
		Columns("source, destination, original, translation").
		Values(t.Source, t.Destination, t.Original, t.Translation).
		ToSql()
	if err != nil {
		r.logger.Error("TranslationRepo.Store - r.Builder", zap.Error(err))
		return fmt.Errorf("TranslationRepo - Store - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		r.logger.Error("TranslationRepo.Store - r.Pool.Exec", zap.Error(err))
		return fmt.Errorf("TranslationRepo - Store - r.Pool.Exec: %w", err)
	}

	r.logger.Info("TranslationRepo.Store finished")
	return nil
}
