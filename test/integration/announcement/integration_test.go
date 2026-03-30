package announcement

// TODO: This integration test needs rewriting for the DDD architecture.
// The old imports have been removed during the DDD migration:
//   - gct/internal/controller/restapi/v1/announcement -> use gct/internal/announcement/interfaces/http
//   - gct/internal/domain -> use gct/internal/announcement/domain (or gct/internal/shared/domain)
//   - gct/internal/repo -> repos are now per-BC under gct/internal/announcement/infrastructure/postgres
//   - gct/internal/usecase -> use cases are now command/query handlers in gct/internal/announcement/application
//
// To rewrite:
//   - Create announcement BC via announcement.NewBoundedContext(pool, eventBus, logger)
//   - Create HTTP handler via announcementhttp.NewHandler(bc, logger)
//   - Call handler methods (Create, Get, List, Update, Delete) directly on the DDD handler
//   - Replace domain.CreateAnnouncementRequest with the DDD command types
//   - See test/integration/user/ddd/ for a working example of DDD integration tests
