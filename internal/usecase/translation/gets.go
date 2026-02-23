package translation

import (
	"context"

	"gct/internal/domain"
)

// Gets returns all translations for an entity reshaped as lang_code → data map.
func (uc *UseCase) Gets(ctx context.Context, filter domain.TranslationFilter) (domain.EntityTranslations, error) {
	rows, err := uc.repo.Gets(ctx, filter)
	if err != nil {
		return nil, err
	}

	result := make(domain.EntityTranslations, len(rows))
	for _, t := range rows {
		result[t.LangCode] = t.Data
	}
	return result, nil
}
