// Package port exposes this BC's Open Host Service adapters: concrete
// implementations of interfaces declared in gct/internal/contract/ports.
// Consumer BCs depend only on the contracts package; the composition root
// wires these adapters in.
package port

import (
	"context"
	"fmt"

	"gct/internal/context/iam/user/application/query"
	"gct/internal/context/iam/user/domain"
	"gct/internal/contract/ports"
	shared "gct/internal/kernel/domain"

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
	u, err := a.h.Handle(ctx, query.FindUserForAuthQuery{UserID: domain.UserID(userID)})
	if err != nil {
		return nil, fmt.Errorf("user.port.AuthLookupAdapter.FindForAuth: %w", err)
	}
	return u, nil
}

// Compile-time assertion that the adapter satisfies the port contract.
var _ ports.AuthUserLookup = (*AuthLookupAdapter)(nil)
