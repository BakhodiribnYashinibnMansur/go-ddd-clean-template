package webhook

// TODO: This integration test needs rewriting for the DDD architecture.
// The old imports have been removed during the DDD migration:
//   - gct/internal/controller/restapi/v1/webhook -> use gct/internal/webhook/interfaces/http
//   - gct/internal/domain -> use gct/internal/webhook/domain (or gct/internal/shared/domain)
//   - gct/internal/repo -> repos are now per-BC under gct/internal/webhook/infrastructure/postgres
//   - gct/internal/usecase -> use cases are now command/query handlers in gct/internal/webhook/application
//
// To rewrite:
//   - Create webhook BC via webhook.NewBoundedContext(pool, eventBus, logger)
//   - Create HTTP handler via webhookhttp.NewHandler(bc, logger)
//   - Call handler methods (Create, Get, List, Update, Delete) on the DDD handler
//   - Replace domain.CreateWebhookRequest with the DDD command types
//   - See test/integration/user/ddd/ for a working example of DDD integration tests
