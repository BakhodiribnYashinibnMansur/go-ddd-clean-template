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

	if u.Attributes == nil {
		u.Attributes = make(map[string]any)
	}

	// Validate input
	if err := validator.ValidateStruct(u); err != nil {
		return err
	}

	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	uc.logger.Infoc(ctx, "user create started", "input", u)

	err := uc.repo.Postgres.User.Client.Create(ctx, u)
	if err != nil {
		uc.logger.Errorc(ctx, "user create failed", "error", err)
		return apperrors.MapRepoToServiceError(err).WithInput(u)
	}

	uc.logger.Infoc(ctx, "user create success")
	return nil
}
