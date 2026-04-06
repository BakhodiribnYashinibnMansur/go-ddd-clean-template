package query

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"
	"gct/internal/kernel/infrastructure/logger"

	"gct/internal/context/iam/generic/authz/application/dto"
	authzrepo "gct/internal/context/iam/generic/authz/domain/repository"
	shared "gct/internal/kernel/domain"
	"gct/internal/kernel/infrastructure/pgxutil"
)

// ListScopesQuery holds the input for listing scopes.
type ListScopesQuery struct {
	Pagination shared.Pagination
}

// ListScopesResult holds the output of the list scopes query.
type ListScopesResult struct {
	Scopes []*dto.ScopeView
	Total  int64
}

// ListScopesHandler handles the ListScopesQuery.
type ListScopesHandler struct {
	readRepo authzrepo.AuthzReadRepository
	logger   logger.Log
}

// NewListScopesHandler creates a new ListScopesHandler.
func NewListScopesHandler(readRepo authzrepo.AuthzReadRepository, l logger.Log) *ListScopesHandler {
	return &ListScopesHandler{readRepo: readRepo, logger: l}
}

// Handle executes the ListScopesQuery and returns a list of ScopeView.
func (h *ListScopesHandler) Handle(ctx context.Context, q ListScopesQuery) (_ *ListScopesResult, err error) {
	ctx, end := pgxutil.AppSpan(ctx, "ListScopesHandler.Handle")
	defer func() { end(err) }()
	defer logger.SlowOp(h.logger, ctx, "ListScopes", "scope")()

	views, total, err := h.readRepo.ListScopes(ctx, q.Pagination)
	if err != nil {
		h.logger.Warnc(ctx, "query failed", logger.F{Op: "ListScopes", Entity: "access", Err: err}.KV()...)
		return nil, apperrors.MapToServiceError(err)
	}

	result := make([]*dto.ScopeView, len(views))
	for i, v := range views {
		result[i] = &dto.ScopeView{
			Path:   v.Path,
			Method: v.Method,
		}
	}

	return &ListScopesResult{
		Scopes: result,
		Total:  total,
	}, nil
}
