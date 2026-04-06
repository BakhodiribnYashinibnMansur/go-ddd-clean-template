package command

import (
	"context"
	"errors"
	"testing"

	translationentity "gct/internal/context/content/generic/translation/domain/entity"

	"github.com/stretchr/testify/require"
)

func TestDeleteTranslationHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockRepo{}
	log := &mockLogger{}

	handler := NewDeleteTranslationHandler(repo, log)

	err := handler.Handle(context.Background(), DeleteTranslationCommand{
		ID: translationentity.NewTranslationID(),
	})
	require.NoError(t, err)
}

func TestDeleteTranslationHandler_RepoError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("repo delete failed")
	errR := &errorRepo{deleteErr: repoErr}
	log := &mockLogger{}

	handler := NewDeleteTranslationHandler(errR, log)

	err := handler.Handle(context.Background(), DeleteTranslationCommand{
		ID: translationentity.NewTranslationID(),
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo delete error, got: %v", err)
	}
}
