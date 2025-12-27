package translation

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/evrone/go-clean-template/internal/repo"
)

// UseCase -.
type UseCase struct {
	repo   repo.TranslationRepo
	webAPI repo.TranslationWebAPI
}

// New -.
func New(r repo.TranslationRepo, w repo.TranslationWebAPI) *UseCase {
	return &UseCase{
		repo:   r,
		webAPI: w,
	}
}

// History - getting translate history from store.
func (uc *UseCase) History(ctx context.Context) (domain.TranslationHistory, error) {
	translations, err := uc.repo.GetHistory(ctx)
	if err != nil {
		return domain.TranslationHistory{}, fmt.Errorf("TranslationUseCase - History - s.repo.GetHistory: %w", err)
	}

	return domain.TranslationHistory{History: translations}, nil
}

// Translate -.
func (uc *UseCase) Translate(ctx context.Context, t domain.Translation) (domain.Translation, error) {
	translation, err := uc.webAPI.Translate(t)
	if err != nil {
		return domain.Translation{}, fmt.Errorf("TranslationUseCase - Translate - s.webAPI.Translate: %w", err)
	}

	err = uc.repo.Store(ctx, translation)
	if err != nil {
		return domain.Translation{}, fmt.Errorf("TranslationUseCase - Translate - s.repo.Store: %w", err)
	}

	return translation, nil
}
