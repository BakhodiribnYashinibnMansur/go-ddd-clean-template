package systemerror

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (uc *UseCase) Resolve(ctx context.Context, id string, resolvedBy *string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid id: %w", err)
	}

	var resolvedByUID *uuid.UUID
	if resolvedBy != nil {
		parsed, err := uuid.Parse(*resolvedBy)
		if err == nil {
			resolvedByUID = &parsed
		}
	}

	return uc.repo.Postgres.Audit.SystemError.Resolve(ctx, uid, resolvedByUID)
}
