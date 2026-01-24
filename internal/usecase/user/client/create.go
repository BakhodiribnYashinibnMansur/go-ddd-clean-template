package client

import (
	"context"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"
	"gct/pkg/validator"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)

func (uc *UseCase) Create(ctx context.Context, u *domain.User) error {
	ctx, span := otel.Tracer("user-client-usecase").Start(ctx, "Create")
	defer span.End()

	// Validate input
	if err := validator.ValidateStruct(ctx, u); err != nil {
		return err
	}

	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	uc.logger.WithContext(ctx).Infow("user create started", "input", u)

	err := uc.repo.Postgres.User.Client.Create(ctx, u)
	if err != nil {
		uc.logger.WithContext(ctx).Errorw("user create failed", "error", err)
		return apperrors.MapRepoToServiceError(ctx, err).WithInput(u)
	}

	uc.logger.WithContext(ctx).Infow("user create success")
	return nil
}
