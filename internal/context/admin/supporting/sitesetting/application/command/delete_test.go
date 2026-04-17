package command

import (
	"context"
	"errors"
	"testing"

	siteentity "gct/internal/context/admin/supporting/sitesetting/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDeleteSiteSettingHandler_Handle(t *testing.T) {
	t.Parallel()

	repo := &mockRepo{}
	log := &mockLogger{}

	handler := NewDeleteSiteSettingHandler(repo, nil, log)

	err := handler.Handle(context.Background(), DeleteSiteSettingCommand{
		ID: siteentity.SiteSettingID(uuid.New()),
	})
	require.NoError(t, err)
}

func TestDeleteSiteSettingHandler_RepoError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("repo delete failed")
	errR := &errorRepo{deleteErr: repoErr}
	log := &mockLogger{}

	handler := NewDeleteSiteSettingHandler(errR, nil, log)

	err := handler.Handle(context.Background(), DeleteSiteSettingCommand{
		ID: siteentity.SiteSettingID(uuid.New()),
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo delete error, got: %v", err)
	}
}
