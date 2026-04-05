# Authz

Bounded context for Role-Based Access Control (RBAC) with Attribute-Based Access Control (ABAC) policy support. Manages roles, permissions, policies, and scopes (API endpoints), along with their relationships.

## Domain

### Aggregate Root
- `Role` -- Authorization role with fields: `name`, optional `description`, and a collection of `Permission` child entities. Supports adding/removing permissions.

### Entities
- `Permission` -- Child entity of Role. Has `name`, optional `parentID` (for hierarchical permissions), optional `description`, and a collection of `Scope` value objects. Supports hierarchical nesting via `parentID`.
- `Policy` -- ABAC policy entity linked to a permission via `permissionID`. Has `effect` (ALLOW/DENY), `priority` (int), `active` (bool), and `conditions` (map[string]any for arbitrary ABAC conditions). Supports toggling active state.

### Value Objects
- `Scope` -- Represents an API endpoint with `Path` and `Method` (GET, POST, PUT, DELETE, PATCH).
- `PolicyEffect` -- String type with constants `PolicyAllow` ("ALLOW") and `PolicyDeny` ("DENY").

### Domain Events
- `RoleCreated` -- Raised when a new role is created. Carries role name.
- `RoleDeleted` -- Raised when a role is deleted.
- `PolicyUpdated` -- Raised when a policy is updated. Carries the policy ID.
- `PermissionGranted` -- Raised when a permission is assigned to a role. Carries the permission ID.

### Domain Errors
- `ErrRoleNotFound` -- Role does not exist.
- `ErrPermissionNotFound` -- Permission does not exist.
- `ErrPolicyNotFound` -- Policy does not exist.
- `ErrScopeNotFound` -- Scope does not exist.
- `ErrDuplicatePermission` -- Permission already exists.

### Repository Interfaces
- `RoleRepository` (write) -- `Save`, `FindByID`, `Update`, `Delete`, `List`
- `PermissionRepository` (write) -- `Save`, `FindByID`, `Update`, `Delete`, `List`
- `PolicyRepository` (write) -- `Save`, `FindByID`, `Update`, `Delete`, `List`, `FindByPermissionID`
- `ScopeRepository` (write) -- `Save`, `Delete` (by path+method), `List`
- `RolePermissionRepository` (write) -- `Assign`, `Revoke` (manages role-permission join table)
- `PermissionScopeRepository` (write) -- `Assign`, `Revoke` (manages permission-scope join table)
- `AuthzReadRepository` (read) -- `GetRole`, `ListRoles`, `GetPermission`, `ListPermissions`, `ListPolicies`, `ListScopes`

## Application (CQRS)

### Commands
- `CreateRoleCommand` / `CreateRoleHandler` -- Creates a new role with name and optional description, publishes `RoleCreated` event.
- `UpdateRoleCommand` / `UpdateRoleHandler` -- Partially updates a role's name and/or description.
- `DeleteRoleCommand` / `DeleteRoleHandler` -- Deletes a role and publishes `RoleDeleted` event.
- `CreatePermissionCommand` / `CreatePermissionHandler` -- Creates a new permission with name, optional parent ID, and description.
- `DeletePermissionCommand` / `DeletePermissionHandler` -- Deletes a permission by ID.
- `CreatePolicyCommand` / `CreatePolicyHandler` -- Creates a new ABAC policy linked to a permission with effect, priority, and conditions.
- `UpdatePolicyCommand` / `UpdatePolicyHandler` -- Partially updates a policy's effect, priority, and/or conditions.
- `DeletePolicyCommand` / `DeletePolicyHandler` -- Deletes a policy by ID.
- `TogglePolicyCommand` / `TogglePolicyHandler` -- Flips a policy's active/inactive state.
- `CreateScopeCommand` / `CreateScopeHandler` -- Registers a new API scope (path + HTTP method).
- `DeleteScopeCommand` / `DeleteScopeHandler` -- Removes a scope by path and method.
- `AssignPermissionCommand` / `AssignPermissionHandler` -- Assigns a permission to a role, publishes `PermissionGranted` event.
- `AssignScopeCommand` / `AssignScopeHandler` -- Assigns a scope to a permission.

### Queries
- `GetRoleQuery` / `GetRoleHandler` -- Fetches a single role view by ID.
- `ListRolesQuery` / `ListRolesHandler` -- Returns a paginated list of role views.
- `ListPermissionsQuery` / `ListPermissionsHandler` -- Returns a paginated list of permission views.
- `ListPoliciesQuery` / `ListPoliciesHandler` -- Returns a paginated list of policy views.
- `ListScopesQuery` / `ListScopesHandler` -- Returns a paginated list of scope views.

## HTTP API

### Roles

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/roles` | Create a new role |
| GET | `/roles` | List roles (paginated) |
| GET | `/roles/:id` | Get a single role by ID |
| PATCH | `/roles/:id` | Update a role |
| DELETE | `/roles/:id` | Delete a role |
| POST | `/roles/:id/permissions` | Assign a permission to a role |

### Permissions

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/permissions` | Create a new permission |
| GET | `/permissions` | List permissions (paginated) |
| DELETE | `/permissions/:id` | Delete a permission |
| POST | `/permissions/:id/scopes` | Assign a scope to a permission |

### Policies

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/policies` | Create a new policy |
| GET | `/policies` | List policies (paginated) |
| PATCH | `/policies/:id` | Update a policy |
| DELETE | `/policies/:id` | Delete a policy |
| POST | `/policies/:id/toggle` | Toggle a policy's active state |

### Scopes

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/scopes` | Create a new scope |
| GET | `/scopes` | List scopes (paginated) |
| DELETE | `/scopes` | Delete a scope (path + method in request body) |

## Usage
```go
import "gct/internal/authz"
```
