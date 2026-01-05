![Go Clean Template](docs/img/logo.svg)

# Go Clean Template

[🇨🇳 中文](README_CN.md) | [🇷🇺 RU](README_RU.md) | [🇺🇿 UZ](PACKAGES_UZ.md)

Clean Architecture template for Golang services.

## Documentation Index

Explore the detailed documentation for each part of the project:

- **[📂 cmd/](cmd/README.md)**: Entry points (Main servers, Migrations, Seeders).
- **[⚙️ config/](config/README.md)**: Configuration management (Env vars).
- **[🧠 internal/domain/](internal/domain/README.md)**: Enterprise Entities (User, Session).
- **[💼 internal/usecase/](internal/usecase/README.md)**: Business Logic and Services.
- **[💾 internal/repo/](internal/repo/README.md)**: Data Access Layer (Postgres, Redis, MinIO).
- **[🌐 internal/controller/](internal/controller/README.md)**: API Handlers (REST, gRPC).
- **[📦 pkg/](pkg/README.md)**: Shared Libraries and Utilities (Logger, Validator).

---

## Overview

The purpose of this template is to demonstrate:
- How to organize a project to prevent "spaghetti code".
- Where to store business logic so it remains independent, clean, and extensible.
- How to keep control as a microservice grows.

It follows the principles of **Clean Architecture** by Robert "Uncle Bob" Martin.

## Project Structure

The application is divided into layers:

1.  **Domain**: Core models. No dependencies.
2.  **UseCase**: Business logic. Depends only on Domain.
3.  **Controllers/Repositories**: External interfaces. Depend on UseCase.

### Data Flow Example
`User Request` -> `Controller` -> `UseCase` -> `Repository` -> `Database`

## Quick Start
### Local Development

```sh
# Start dependencies (Postgres, Redis, etc.)
make compose-up

# Run the application
make run
```

### Integration Tests
```sh
make compose-up-integration-test
```
