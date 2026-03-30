package session

// TODO: This integration test needs rewriting for the DDD architecture.
// The old imports have been removed during the DDD migration:
//   - gct/internal/controller/restapi/v1/user/session -> use gct/internal/session/interfaces/http
//   - gct/internal/domain -> use gct/internal/session/domain (or gct/internal/shared/domain)
//   - gct/internal/repo -> repos are now per-BC under gct/internal/session/infrastructure/postgres
//   - gct/internal/usecase -> use cases are now command/query handlers in gct/internal/session/application
//
// To rewrite:
//   - Create session BC via session.NewBoundedContext(pool, logger) (note: no eventBus for read-only BCs)
//   - Create HTTP handler via sessionhttp.NewHandler(bc, logger)
//   - Call handler methods (List, Get) on the DDD handler
//   - The old Sessions/Session/UpdateActivity/Delete/RevokeCurrent/RevokeAll/Create methods
//     may need new DDD commands or may map to the session BC's query handlers
//   - See test/integration/user/ddd/ for a working example of DDD integration tests
