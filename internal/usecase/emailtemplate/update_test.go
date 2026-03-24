package emailtemplate_test

import (
	"errors"
	"testing"
	"time"

	"gct/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdate_Success_AllFields(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	existing := &domain.EmailTemplate{
		ID:        "tmpl-1",
		Name:      "Old Name",
		Subject:   "Old Subject",
		HtmlBody:  "<p>Old</p>",
		TextBody:  "Old",
		Type:      "transactional",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	newName := "New Name"
	newSubject := "New Subject"
	newHtml := "<p>New</p>"
	newText := "New"
	newType := "marketing"
	isActive := false

	req := domain.UpdateEmailTemplateRequest{
		Name:     &newName,
		Subject:  &newSubject,
		HtmlBody: &newHtml,
		TextBody: &newText,
		Type:     &newType,
		IsActive: &isActive,
	}

	repo.On("GetByID", ctx, "tmpl-1").Return(existing, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*domain.EmailTemplate")).Return(nil)

	result, err := uc.Update(ctx, "tmpl-1", req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, newName, result.Name)
	assert.Equal(t, newSubject, result.Subject)
	assert.Equal(t, newHtml, result.HtmlBody)
	assert.Equal(t, newText, result.TextBody)
	assert.Equal(t, newType, result.Type)
	assert.False(t, result.IsActive)
	repo.AssertExpectations(t)
}

func TestUpdate_PartialFields(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	existing := &domain.EmailTemplate{
		ID:       "tmpl-2",
		Name:     "Original",
		Subject:  "Original Sub",
		HtmlBody: "<p>Original</p>",
		TextBody: "Original",
		Type:     "transactional",
		IsActive: true,
	}

	newName := "Updated Name"
	req := domain.UpdateEmailTemplateRequest{
		Name: &newName,
	}

	repo.On("GetByID", ctx, "tmpl-2").Return(existing, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*domain.EmailTemplate")).Return(nil)

	result, err := uc.Update(ctx, "tmpl-2", req)

	require.NoError(t, err)
	assert.Equal(t, newName, result.Name)
	assert.Equal(t, "Original Sub", result.Subject)
	assert.Equal(t, "<p>Original</p>", result.HtmlBody)
	assert.Equal(t, "Original", result.TextBody)
	assert.Equal(t, "transactional", result.Type)
	assert.True(t, result.IsActive)
	repo.AssertExpectations(t)
}

func TestUpdate_GetByIDError(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	req := domain.UpdateEmailTemplateRequest{}

	repo.On("GetByID", ctx, "nonexistent").
		Return(nil, errors.New("not found"))

	result, err := uc.Update(ctx, "nonexistent", req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
	repo.AssertExpectations(t)
}

func TestUpdate_RepoUpdateError(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	existing := &domain.EmailTemplate{
		ID:   "tmpl-3",
		Name: "Test",
	}

	newName := "Updated"
	req := domain.UpdateEmailTemplateRequest{
		Name: &newName,
	}

	repo.On("GetByID", ctx, "tmpl-3").Return(existing, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*domain.EmailTemplate")).
		Return(errors.New("update failed"))

	result, err := uc.Update(ctx, "tmpl-3", req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "update failed")
	repo.AssertExpectations(t)
}
