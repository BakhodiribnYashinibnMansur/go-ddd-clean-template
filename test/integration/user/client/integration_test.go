package client

// TODO: This integration test needs rewriting for the DDD architecture.
// The old imports have been removed during the DDD migration:
//   - gct/internal/controller/restapi/v1/authz/auth -> use gct/internal/user/interfaces/http (SignIn/SignUp/SignOut)
//   - gct/internal/controller/restapi/v1/user/client -> use gct/internal/user/interfaces/http (CRUD)
//   - gct/internal/domain -> use gct/internal/user/domain (or gct/internal/shared/domain)
//   - gct/internal/repo -> repos are now per-BC under gct/internal/user/infrastructure/postgres
//   - gct/internal/usecase -> use cases are now command/query handlers in gct/internal/user/application
//
// To rewrite:
//   - Create user BC via user.NewBoundedContext(pool, eventBus, logger, jwtCfg)
//   - Create HTTP handler via userhttp.NewHandler(bc, logger)
//   - Auth methods: handler.SignIn, handler.SignUp, handler.SignOut
//   - CRUD methods: handler.Create, handler.Get, handler.List, handler.Update, handler.Delete
//   - Replace domain.NewUser / domain.UserFilter with DDD domain/command types
//   - See test/integration/user/ddd/ for a working example of DDD integration tests
