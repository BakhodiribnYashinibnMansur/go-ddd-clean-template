package translation

import (
	"context"
	"fmt"

	"gct/internal/domain"

	"github.com/google/uuid"
)

// Upsert saves translations for all languages in the request.
// Existing fields are merged (JSONB ||), new ones are added.
func (uc *UseCase) Upsert(ctx context.Context, entityType string, entityID uuid.UUID, req domain.UpsertTranslationsRequest) error {
	for langCode, data := range req {
		if err := uc.repo.Upsert(ctx, entityType, entityID, langCode, data); err != nil {
			return fmt.Errorf("translation upsert [%s]: %w", langCode, err)
		}
	}
	return nil
}
