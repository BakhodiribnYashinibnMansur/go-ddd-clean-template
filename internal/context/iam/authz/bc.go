package authz

import (
	"gct/internal/context/iam/authz/application/command"
	"gct/internal/context/iam/authz/application/query"
	"gct/internal/context/iam/authz/infrastructure/postgres"
	"gct/internal/platform/application"
	"gct/internal/platform/infrastructure/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// BoundedContext wires together all command and query handlers for the Authz BC.
type BoundedContext struct {
	// Commands — Roles
	CreateRole *command.CreateRoleHandler
	UpdateRole *command.UpdateRoleHandler
	DeleteRole *command.DeleteRoleHandler

	// Commands — Permissions
	CreatePermission *command.CreatePermissionHandler
	DeletePermission *command.DeletePermissionHandler

	// Commands — Policies
	CreatePolicy *command.CreatePolicyHandler
	UpdatePolicy *command.UpdatePolicyHandler
	DeletePolicy *command.DeletePolicyHandler
	TogglePolicy *command.TogglePolicyHandler

	// Commands — Scopes
	CreateScope *command.CreateScopeHandler
	DeleteScope *command.DeleteScopeHandler

	// Commands — Assignments
	AssignPermission *command.AssignPermissionHandler
	AssignScope      *command.AssignScopeHandler

	// Queries
	GetRole         *query.GetRoleHandler
	ListRoles       *query.ListRolesHandler
	ListPermissions *query.ListPermissionsHandler
	ListPolicies    *query.ListPoliciesHandler
	ListScopes      *query.ListScopesHandler
	CheckAccess     *query.CheckAccessHandler
}

// NewBoundedContext creates a fully wired Authz bounded context.
func NewBoundedContext(pool *pgxpool.Pool, eventBus application.EventBus, l logger.Log) *BoundedContext {
	roleRepo := postgres.NewRoleWriteRepo(pool)
	permRepo := postgres.NewPermissionWriteRepo(pool)
	policyRepo := postgres.NewPolicyWriteRepo(pool)
	scopeRepo := postgres.NewScopeWriteRepo(pool)
	rolePermRepo := postgres.NewRolePermissionRepo(pool)
	permScopeRepo := postgres.NewPermissionScopeRepo(pool)
	readRepo := postgres.NewAuthzReadRepo(pool)

	return &BoundedContext{
		// Commands — Roles
		CreateRole: command.NewCreateRoleHandler(roleRepo, eventBus, l),
		UpdateRole: command.NewUpdateRoleHandler(roleRepo, eventBus, l),
		DeleteRole: command.NewDeleteRoleHandler(roleRepo, eventBus, l),

		// Commands — Permissions
		CreatePermission: command.NewCreatePermissionHandler(permRepo, l),
		DeletePermission: command.NewDeletePermissionHandler(permRepo, l),

		// Commands — Policies
		CreatePolicy: command.NewCreatePolicyHandler(policyRepo, l),
		UpdatePolicy: command.NewUpdatePolicyHandler(policyRepo, l),
		DeletePolicy: command.NewDeletePolicyHandler(policyRepo, l),
		TogglePolicy: command.NewTogglePolicyHandler(policyRepo, l),

		// Commands — Scopes
		CreateScope: command.NewCreateScopeHandler(scopeRepo, l),
		DeleteScope: command.NewDeleteScopeHandler(scopeRepo, l),

		// Commands — Assignments
		AssignPermission: command.NewAssignPermissionHandler(rolePermRepo, eventBus, l),
		AssignScope:      command.NewAssignScopeHandler(permScopeRepo, l),

		// Queries
		GetRole:         query.NewGetRoleHandler(readRepo, l),
		ListRoles:       query.NewListRolesHandler(readRepo, l),
		ListPermissions: query.NewListPermissionsHandler(readRepo, l),
		ListPolicies:    query.NewListPoliciesHandler(readRepo, l),
		ListScopes:      query.NewListScopesHandler(readRepo, l),
		CheckAccess:     query.NewCheckAccessHandler(readRepo, l),
	}
}
