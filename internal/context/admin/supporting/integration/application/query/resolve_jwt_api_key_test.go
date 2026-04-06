package query

import (
	"context"
	"errors"
	"testing"
	"time"

	integentity "gct/internal/context/admin/supporting/integration/domain/entity"
	"gct/internal/kernel/infrastructure/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// jwtResolverRepo is a recording mock implementing IntegrationReadRepository
// with a programmable response for FindJWTByHash.
type jwtResolverRepo struct {
	view      *integentity.JWTIntegrationView
	findErr   error
	findCalls int
}

func (r *jwtResolverRepo) FindByID(_ context.Context, _ integentity.IntegrationID) (*integentity.IntegrationView, error) {
	return nil, integentity.ErrIntegrationNotFound
}
func (r *jwtResolverRepo) List(_ context.Context, _ integentity.IntegrationFilter) ([]*integentity.IntegrationView, int64, error) {
	return nil, 0, nil
}
func (r *jwtResolverRepo) FindByAPIKey(_ context.Context, _ string) (*integentity.IntegrationAPIKeyView, error) {
	return nil, integentity.ErrIntegrationNotFound
}
func (r *jwtResolverRepo) ListActiveJWT(_ context.Context) ([]integentity.JWTIntegrationView, error) {
	return nil, nil
}
func (r *jwtResolverRepo) FindJWTByHash(_ context.Context, _ []byte) (*integentity.JWTIntegrationView, error) {
	r.findCalls++
	if r.findErr != nil {
		return nil, r.findErr
	}
	return r.view, nil
}

const (
	validPlaintext   = "0123456789abcdef0123456789abcdef0123" // 36 chars
	anotherPlaintext = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // 36 chars
)

func TestResolveJWTAPIKeyHandler_HappyPath(t *testing.T) {
	t.Parallel()
	expected := &integentity.JWTIntegrationView{
		ID:          integentity.NewIntegrationID(),
		Name:        "stripe",
		BindingMode: integentity.BindingModeWarn,
	}
	repo := &jwtResolverRepo{view: expected}
	h := NewResolveJWTAPIKeyHandler(repo, []byte("pepper-secret-000000"), 30*time.Second, logger.Noop())

	got, err := h.Handle(context.Background(), ResolveJWTAPIKeyQuery{PlainAPIKey: validPlaintext})
	require.NoError(t, err)
	assert.Equal(t, expected, got)
	assert.Equal(t, 1, repo.findCalls)
}

func TestResolveJWTAPIKeyHandler_CacheHit(t *testing.T) {
	t.Parallel()
	expected := &integentity.JWTIntegrationView{ID: integentity.NewIntegrationID(), Name: "cached"}
	repo := &jwtResolverRepo{view: expected}
	h := NewResolveJWTAPIKeyHandler(repo, []byte("pepper-secret-000000"), 30*time.Second, logger.Noop())

	_, err := h.Handle(context.Background(), ResolveJWTAPIKeyQuery{PlainAPIKey: validPlaintext})
	require.NoError(t, err)
	_, err = h.Handle(context.Background(), ResolveJWTAPIKeyQuery{PlainAPIKey: validPlaintext})
	require.NoError(t, err)

	assert.Equal(t, 1, repo.findCalls, "second call should hit cache, not repo")
}

func TestResolveJWTAPIKeyHandler_Invalidate(t *testing.T) {
	t.Parallel()
	expected := &integentity.JWTIntegrationView{ID: integentity.NewIntegrationID(), Name: "ephemeral"}
	repo := &jwtResolverRepo{view: expected}
	h := NewResolveJWTAPIKeyHandler(repo, []byte("pepper-secret-000000"), 30*time.Second, logger.Noop())

	_, _ = h.Handle(context.Background(), ResolveJWTAPIKeyQuery{PlainAPIKey: validPlaintext})
	h.Invalidate(validPlaintext)
	_, _ = h.Handle(context.Background(), ResolveJWTAPIKeyQuery{PlainAPIKey: validPlaintext})

	assert.Equal(t, 2, repo.findCalls, "invalidation should drop cache entry")
}

func TestResolveJWTAPIKeyHandler_WrongKeyReturnsNotFound(t *testing.T) {
	t.Parallel()
	repo := &jwtResolverRepo{findErr: integentity.ErrIntegrationNotFound}
	h := NewResolveJWTAPIKeyHandler(repo, []byte("pepper-secret-000000"), 30*time.Second, logger.Noop())

	got, err := h.Handle(context.Background(), ResolveJWTAPIKeyQuery{PlainAPIKey: anotherPlaintext})
	require.Error(t, err)
	require.True(t, errors.Is(err, integentity.ErrAPIKeyNotFound))
	assert.Nil(t, got)
}

func TestResolveJWTAPIKeyHandler_RejectsShortKey(t *testing.T) {
	t.Parallel()
	repo := &jwtResolverRepo{}
	h := NewResolveJWTAPIKeyHandler(repo, []byte("pepper-secret-000000"), 30*time.Second, logger.Noop())

	_, err := h.Handle(context.Background(), ResolveJWTAPIKeyQuery{PlainAPIKey: "too-short"})
	require.Error(t, err)
	require.True(t, errors.Is(err, integentity.ErrInvalidJWTAPIKey))
	assert.Equal(t, 0, repo.findCalls, "should not hit repo for invalid input")
}
