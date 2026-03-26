# Plan 2: Core Bounded Contexts — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development

**Goal:** Create User, Session, and Authz bounded contexts with full DDD structure (domain, application, infrastructure, interfaces layers)

**Approach:** Build new DDD code alongside existing code. Old code stays untouched and functional. Rewiring happens in Plan 6.

---

## Task 1: User BC — Domain Layer

**Files to create:**
- `internal/user/domain/entity.go` — User Aggregate Root (private fields, embeds AggregateRoot)
- `internal/user/domain/session.go` — Session child entity (within User aggregate)
- `internal/user/domain/value_object.go` — Phone, Email, Password Value Objects (self-validating)
- `internal/user/domain/repository.go` — UserRepository interface (extends Generic)
- `internal/user/domain/event.go` — UserCreated, UserSignedIn, UserDeactivated events
- `internal/user/domain/error.go` — Domain errors
- `internal/user/domain/service.go` — SignInService, PasswordService domain services
- Tests for all files

## Task 2: User BC — Application Layer (Commands)

**Files to create:**
- `internal/user/application/command/create_user.go`
- `internal/user/application/command/update_user.go`
- `internal/user/application/command/delete_user.go`
- `internal/user/application/command/sign_in.go`
- `internal/user/application/command/sign_up.go`
- `internal/user/application/command/sign_out.go`
- `internal/user/application/command/approve_user.go`
- `internal/user/application/command/change_role.go`
- `internal/user/application/command/bulk_action.go`
- `internal/user/application/dto.go` — UserView, response DTOs
- Tests for handlers

## Task 3: User BC — Application Layer (Queries)

**Files to create:**
- `internal/user/application/query/get_user.go`
- `internal/user/application/query/list_users.go`
- Tests

## Task 4: User BC — Infrastructure (Postgres repos)

**Files to create:**
- `internal/user/infrastructure/postgres/write_repo.go`
- `internal/user/infrastructure/postgres/read_repo.go`
- `internal/user/bc.go` — BoundedContext factory

## Task 5: Authz BC — Domain + Application + Infrastructure

**Files to create:**
- `internal/authz/domain/` — Role (aggregate root), Permission, Policy, Scope, Relation entities
- `internal/authz/application/command/` — CRUD handlers for role, permission, policy, scope
- `internal/authz/application/query/` — Get/List handlers
- `internal/authz/infrastructure/postgres/` — Write/Read repos
- `internal/authz/bc.go`

## Task 6: Session BC — Read-only Query Layer

**Files to create:**
- `internal/session/application/query/` — GetSession, ListSessions
- `internal/session/infrastructure/postgres/read_repo.go`
- `internal/session/bc.go`
