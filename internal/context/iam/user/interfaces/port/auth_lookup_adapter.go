// Package port exposes this BC's Open Host Service adapters: concrete
// implementations of interfaces declared in gct/internal/contract/ports.
// Consumer BCs depend only on the contracts package; the composition root
// wires these adapters in.
package port

import (
	"context"

	"gct/internal/context/iam/user/application/query"
	"gct/internal/contract/ports"
	shared "gct/internal/platform/domain"

	"github.com/google/uuid"
)

// AuthLookupAdapter wraps the user BC's FindUserForAuth query handler and
// satisfies contracts/ports.AuthUserLookup so consumers (e.g. authz) can
// resolve a user's auth projection without importing this BC.
type AuthLookupAdapter struct {
	h *query.FindUserForAuthHandler
}

// NewAuthLookupAdapter builds the adapter. It is expected to be constructed
// once in the composition root.
func NewAuthLookupAdapter(h *query.FindUserForAuthHandler) *AuthLookupAdapter {
	return &AuthLookupAdapter{h: h}
}

// FindForAuth resolves the user's minimal auth projection by delegating to
// the query handler.
func (a *AuthLookupAdapter) FindForAuth(ctx context.Context, userID uuid.UUID) (*shared.AuthUser, error) {
	return a.h.Handle(ctx, query.FindUserForAuthQuery{UserID: userID})
}

// Compile-time assertion that the adapter satisfies the port contract.
var _ ports.AuthUserLookup = (*AuthLookupAdapter)(nil)
