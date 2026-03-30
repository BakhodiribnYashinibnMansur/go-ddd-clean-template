package role

// TODO: This integration test needs rewriting for the DDD architecture.
// The old imports have been removed during the DDD migration:
//   - gct/internal/controller/restapi/v1/authz/role -> use gct/internal/authz/interfaces/http
//   - gct/internal/controller/restapi/v1/authz/permission -> use gct/internal/authz/interfaces/http
//   - gct/internal/domain -> use gct/internal/authz/domain (or gct/internal/shared/domain)
//   - gct/internal/repo -> repos are now per-BC under gct/internal/authz/infrastructure/postgres
//   - gct/internal/usecase -> use cases are now command/query handlers in gct/internal/authz/application
//
// To rewrite:
//   - Create authz BC via authz.NewBoundedContext(pool, eventBus, logger)
//   - Create HTTP handler via authzhttp.NewHandler(bc, logger)
//   - Call handler methods (CreateRole, ListRoles, GetRole, UpdateRole, DeleteRole, AssignPermission)
//   - Replace domain.Role / domain.Permission with DDD domain types
//   - See test/integration/user/ddd/ for a working example of DDD integration tests
