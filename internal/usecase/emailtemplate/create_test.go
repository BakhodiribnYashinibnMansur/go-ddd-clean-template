package emailtemplate_test

import (
	"errors"
	"testing"

	"gct/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreate_Success(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	req := domain.CreateEmailTemplateRequest{
		Name:     "Welcome Email",
		Subject:  "Welcome!",
		HtmlBody: "<h1>Hello</h1>",
		TextBody: "Hello",
		Type:     "transactional",
	}

	repo.On("Create", ctx, mock.AnythingOfType("*domain.EmailTemplate")).
		Return(nil)

	result, err := uc.Create(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, req.Name, result.Name)
	assert.Equal(t, req.Subject, result.Subject)
	assert.Equal(t, req.HtmlBody, result.HtmlBody)
	assert.Equal(t, req.TextBody, result.TextBody)
	assert.Equal(t, req.Type, result.Type)
	assert.True(t, result.IsActive)
	repo.AssertExpectations(t)
}

func TestCreate_RepoError(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	req := domain.CreateEmailTemplateRequest{
		Name:     "Fail Template",
		Subject:  "Subject",
		HtmlBody: "<p>Body</p>",
		Type:     "marketing",
	}

	repo.On("Create", ctx, mock.AnythingOfType("*domain.EmailTemplate")).
		Return(errors.New("database error"))

	result, err := uc.Create(ctx, req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
	repo.AssertExpectations(t)
}

func TestCreate_GeneratesUUID(t *testing.T) {
	ctx := t.Context()
	uc, repo := setup(t)

	req := domain.CreateEmailTemplateRequest{
		Name:     "UUID Test",
		Subject:  "Test",
		HtmlBody: "<p>Test</p>",
		Type:     "transactional",
	}

	repo.On("Create", ctx, mock.MatchedBy(func(tmpl *domain.EmailTemplate) bool {
		return tmpl.ID != "" && len(tmpl.ID) == 36 // UUID format
	})).Return(nil)

	result, err := uc.Create(ctx, req)

	require.NoError(t, err)
	assert.Len(t, result.ID, 36)
	repo.AssertExpectations(t)
}
