# User

Core bounded context for user management, authentication, and session lifecycle. Implements a rich domain model with CQRS command/query separation, domain events, and a SignInService domain service.

## Domain

### Aggregate Root
- `User` -- embeds `shared.AggregateRoot`. Key fields: phone (Phone), email (*Email), username (*string), password (Password), roleID (*uuid.UUID), attributes (map[string]any), active (bool), isApproved (bool), lastSeen (*time.Time), sessions ([]Session). Created via `NewUser(phone, password, ...UserOption)` or reconstructed from persistence via `ReconstructUser(...)`. Supports functional options: `WithEmail`, `WithUsername`, `WithRoleID`, `WithAttributes`.

### Child Entity
- `Session` -- embeds `shared.BaseEntity`. Key fields: userID, deviceID, deviceName, deviceType (SessionDeviceType), ipAddress, userAgent, refreshTokenHash, expiresAt, lastActivity, revoked. Default session duration is 7 days. Methods: `IsExpired()`, `IsActive()`, `Revoke()`, `UpdateActivity()`, `SetRefreshTokenHash(hash)`.

### Value Objects
- `Phone` -- validated phone number (non-empty, starts with `+`, min 8 chars)
- `Email` -- validated email address (non-empty, contains `@`)
- `Password` -- bcrypt-hashed password (min 8 chars raw). Methods: `Hash()`, `Compare(raw)`
- `SessionDeviceType` -- string enum: DESKTOP, MOBILE, TABLET, BOT, TV

### Domain Service
- `SignInService` -- stateless service that orchestrates sign-in: checks user is active and approved, verifies password, creates a session on the aggregate, and updates last-seen timestamp

### Domain Events
- `UserCreated` (`user.created`) -- raised when a new user is registered
- `UserSignedIn` (`user.signed_in`) -- raised when a user successfully signs in (includes SessionID, IPAddress)
- `UserDeactivated` (`user.deactivated`) -- raised when a user is deactivated
- `PasswordChanged` (`user.password_changed`) -- raised when a user changes their password
- `UserApproved` (`user.approved`) -- raised when a user is approved
- `RoleChanged` (`user.role_changed`) -- raised when a user's role changes (includes OldRoleID, NewRoleID)

### Domain Errors
- `ErrUserNotFound`, `ErrPhoneExists`, `ErrInvalidPassword`, `ErrUserInactive`, `ErrUserNotApproved`, `ErrMaxSessionsReached` (limit: 10), `ErrSessionNotFound`, `ErrWeakPassword`, `ErrInvalidPhone`, `ErrInvalidEmail`

### Aggregate Behaviour
- `AddSession(deviceType, ip, userAgent)` -- creates session, raises UserSignedIn event
- `RemoveSession(sessionID)` -- removes session by ID
- `RevokeAllSessions()` -- revokes every session
- `VerifyPassword(raw)` / `ChangePassword(old, new)` -- password verification and change with event
- `Activate()` / `Deactivate()` -- toggle active status
- `Approve()` -- marks user as approved
- `ChangeRole(roleID)` -- sets new role with event
- `UpdateLastSeen()` -- sets lastSeen to now

### Repository Interfaces
- `UserRepository` (write) -- extends `shared.Repository[User]` with `FindByPhone(phone)`, `FindByEmail(email)`
- `UserReadRepository` (read) -- `FindByID(id)`, `List(filter)`

## Application (CQRS)

### Commands
- `CreateUserCommand` / `CreateUserHandler` -- creates a new user with phone, password, and optional email/username/roleID/attributes; publishes domain events
- `SignUpCommand` / `SignUpHandler` -- self-registration (active but not approved by default); publishes domain events
- `SignInCommand` / `SignInHandler` -- authenticates via phone or email login, creates session using SignInService; returns SignInResult (UserID, SessionID, AccessToken, RefreshToken)
- `SignOutCommand` / `SignOutHandler` -- removes a specific session from the user aggregate
- `UpdateUserCommand` / `UpdateUserHandler` -- updates email, username, and/or attributes on an existing user
- `DeleteUserCommand` / `DeleteUserHandler` -- soft-deletes and deactivates a user; publishes UserDeactivated event
- `ApproveUserCommand` / `ApproveUserHandler` -- marks a user as approved; publishes UserApproved event
- `ChangeRoleCommand` / `ChangeRoleHandler` -- changes the user's role; publishes RoleChanged event
- `BulkActionCommand` / `BulkActionHandler` -- performs activate/deactivate/delete on a list of user IDs

### Queries
- `GetUserQuery` / `GetUserHandler` -- fetches a single user by ID, returns `UserView`
- `ListUsersQuery` / `ListUsersHandler` -- lists users with optional phone/email/active/isApproved filters and pagination, returns `[]*UserView` with total count

## HTTP API

### User Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /users | Create a new user (admin) |
| GET | /users | List users with filters (phone, email, active, limit, offset) |
| GET | /users/:id | Get a single user by UUID |
| PATCH | /users/:id | Update user fields (email, username, attributes) |
| DELETE | /users/:id | Soft-delete a user |
| POST | /users/:id/approve | Approve a user |
| POST | /users/:id/role | Change a user's role |
| POST | /users/bulk-action | Bulk activate/deactivate/delete users |

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /auth/sign-in | Sign in with login (phone or email) + password + device_type |
| POST | /auth/sign-up | Self-register with phone + password (optional username, email) |
| POST | /auth/sign-out | Sign out by removing a session (user_id + session_id) |

## Wiring

`BoundedContext` struct wires all 9 command handlers and 2 query handlers. Created via `NewBoundedContext(pool, eventBus, logger)`.

## Usage
```go
import "gct/internal/user"
```
