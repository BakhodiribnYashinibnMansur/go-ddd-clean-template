package access

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"gct/consts"
	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/google/uuid"
)

func (u *UseCase) Check(ctx context.Context, userID uuid.UUID, session *domain.Session, path, method string, env map[string]any) (bool, error) {
	u.logger.Infow("access check started", "user_id", userID, "path", path, "method", method)

	// 1. Get User
	user, err := u.repo.Postgres.User.Client.Get(ctx, &domain.UserFilter{ID: &userID})
	if err != nil {
		if errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "context canceled") {
			return false, err
		}
		appErr := apperrors.MapRepoToServiceError(err, apperrors.ErrUserNotFound).WithInput(userID)
		u.logger.Errorw("access check failed: get user", "error", appErr)
		return false, appErr
	}
	if user.RoleID == nil {
		err := apperrors.New(apperrors.ErrServiceRoleNotFound, "user has no role").WithInput(userID)
		u.logger.Warnw("access check denied: user has no role", "user_id", userID)
		return false, err
	}

	// Mock RBAC: Check Role Name
	role, err := u.repo.Postgres.Authz.Role.Get(ctx, &domain.RoleFilter{ID: user.RoleID})
	if err != nil {
		appErr := apperrors.MapRepoToServiceError(err, apperrors.ErrServiceRoleNotFound).WithInput(user.RoleID)
		u.logger.Errorw("access check failed: get role", "error", appErr)
		return false, appErr
	}

	// Allow Admin everywhere
	if strings.Contains(strings.ToLower(role.Name), consts.RoleAdmin) {
		u.logger.Infow("access check allowed: admin role", "role", role.Name)
		u.logAudit(ctx, userID, session, path, method, true, "admin allowed", nil)
		return true, nil
	}

	// 2. Get Policies for Role (ABAC)
	policies, err := u.repo.Postgres.Authz.Policy.GetByRole(ctx, *user.RoleID)
	if err != nil {
		u.logger.Errorw("access check failed: get policies", "error", err)
		// Fallback to strict deny if policy fetch fails? Or continue to pure RBAC?
		// Let's assume strict fail-safe
		return false, nil
	}

	// Iterate policies in priority order
	for _, policy := range policies {
		if !evaluateConditions(policy.Conditions, env) {
			continue // Condition not met, skip this policy
		}

		// Condition met, apply effect
		if policy.Effect == domain.PolicyEffectDeny {
			u.logger.Infow("access check denied: policy deny", "policy_id", policy.ID)
			u.logAudit(ctx, userID, session, path, method, false, "policy deny", &policy.ID)
			return false, nil
		}
		if policy.Effect == domain.PolicyEffectAllow {
			// Explicit allow from policy?
			// Usually in hybrid systems, Policy Allow grants access, OR generic RBAC grants access.
			// But Policy Deny overrides everything.
			u.logger.Infow("access check allowed: policy allow", "policy_id", policy.ID)
			u.logAudit(ctx, userID, session, path, method, true, "policy allow", &policy.ID)
			return true, nil
		}
	}

	// 3. Get Permissions for Role (RBAC)
	perms, err := u.repo.Postgres.Authz.Role.GetPermissions(ctx, *user.RoleID)
	if err != nil {
		u.logger.Errorw("access check failed: get permissions", "error", err)
		// Don't return error to user, just deny access (fail safe)
		return false, nil
	}

	// 4. Check Scopes for each Permission
	for _, perm := range perms {
		scopes, err := u.repo.Postgres.Authz.Permission.GetScopes(ctx, perm.ID)
		if err != nil {
			u.logger.Errorw("access check failed: get scopes", "error", err, "perm_id", perm.ID)
			continue
		}

		for _, scope := range scopes {
			// Check Method
			if scope.Method != method && scope.Method != "*" {
				continue
			}

			// Check Path (Exact match or Prefix match if needed)
			// For simplicity: Exact match or simple wildcard
			if scope.Path == path || scope.Path == "*" {
				u.logger.Infow("access check allowed: permission granted", "perm", perm.Name, "scope", scope.Path)
				u.logAudit(ctx, userID, session, path, method, true, "permission granted: "+perm.Name, nil)
				return true, nil
			}

			// Add basic pattern matching if needed, e.g. /api/v1/users/*
			if strings.HasSuffix(scope.Path, "*") {
				prefix := strings.TrimSuffix(scope.Path, "*")
				if strings.HasPrefix(path, prefix) {
					u.logger.Infow("access check allowed: permission granted (wildcard)", "perm", perm.Name, "scope", scope.Path)
					u.logAudit(ctx, userID, session, path, method, true, "permission granted: "+perm.Name, nil)
					return true, nil
				}
			}
		}
	}

	u.logger.Infow("access check denied: insufficient permissions/policies", "role", role.Name)
	u.logAudit(ctx, userID, session, path, method, false, "insufficient permissions", nil)
	return false, nil
}

func (u *UseCase) CheckBatch(ctx context.Context, userID uuid.UUID, session *domain.Session, targets map[string]string, method string, env map[string]any) (map[string]bool, error) {
	u.logger.Infow("access check batch started", "user_id", userID, "count", len(targets))
	results := make(map[string]bool)

	// 1. Get User
	user, err := u.repo.Postgres.User.Client.Get(ctx, &domain.UserFilter{ID: &userID})
	if err != nil {
		u.logger.Errorw("access check batch failed: get user", "error", err)
		return nil, err
	}
	if user.RoleID == nil {
		u.logger.Warnw("access check batch denied: user has no role", "user_id", userID)
		return nil, nil // Return nil or empty results (all false)
	}

	// 2. Mock RBAC: Check Role Name
	role, err := u.repo.Postgres.Authz.Role.Get(ctx, &domain.RoleFilter{ID: user.RoleID})
	if err != nil {
		u.logger.Errorw("access check batch failed: get role", "error", err)
		return nil, err
	}

	// Admin Access
	if strings.Contains(strings.ToLower(role.Name), consts.RoleAdmin) {
		for k := range targets {
			results[k] = true
		}
		u.logger.Infow("access check batch allowed: admin role", "role", role.Name)
		return results, nil
	}

	// 3. Get Policies
	policies, err := u.repo.Postgres.Authz.Policy.GetByRole(ctx, *user.RoleID)
	if err != nil {
		u.logger.Errorw("access check batch failed: get policies", "error", err)
		return nil, err
	}

	// 4. Get Permissions & Scopes (Optimization: Fetch all ONCE)
	perms, err := u.repo.Postgres.Authz.Role.GetPermissions(ctx, *user.RoleID)
	if err != nil {
		u.logger.Errorw("access check batch failed: get permissions", "error", err)
		return nil, err
	}

	var allScopes []*domain.Scope
	for _, perm := range perms {
		scopes, err := u.repo.Postgres.Authz.Permission.GetScopes(ctx, perm.ID)
		if err == nil {
			allScopes = append(allScopes, scopes...)
		}
	}

	// 5. Evaluate for each target
	for key, path := range targets {
		allowed := false
		policyDeny := false
		policyAllow := false

		// Check Policies
		for _, policy := range policies {
			// Note: env might need path update if policies depend on it.
			// Currently re-using env. To be strictly correct, we should copy env and set "path".
			// But for now assuming policies are mostly Env/User based.
			if !evaluateConditions(policy.Conditions, env) {
				continue
			}
			if policy.Effect == domain.PolicyEffectDeny {
				policyDeny = true
				break
			}
			if policy.Effect == domain.PolicyEffectAllow {
				policyAllow = true
			}
		}

		if policyDeny {
			results[key] = false
			continue
		}
		if policyAllow {
			results[key] = true
			continue
		}

		// Check RBAC Scopes
		for _, scope := range allScopes {
			if scope.Method == method || scope.Method == "*" {
				if scope.Path == path || scope.Path == "*" {
					allowed = true
					break
				}
				if strings.HasSuffix(scope.Path, "*") {
					prefix := strings.TrimSuffix(scope.Path, "*")
					if strings.HasPrefix(path, prefix) {
						allowed = true
						break
					}
				}
			}
		}
		results[key] = allowed
	}

	u.logger.Infow("access check batch completed")
	return results, nil
}

func (u *UseCase) logAudit(_ context.Context, userID uuid.UUID, session *domain.Session, path, method string, success bool, decision string, policyID *uuid.UUID) {
	al := &domain.AuditLog{
		ID:        uuid.New(),
		UserID:    &userID,
		Action:    domain.AuditActionPolicyEvaluated,
		Platform:  nil, // Can be inferred from UA if needed
		Decision:  &decision,
		PolicyID:  policyID,
		Success:   success,
		CreatedAt: time.Now(),
		Metadata: map[string]any{
			"path":   path,
			"method": method,
		},
	}

	if session != nil {
		al.SessionID = &session.ID
	}

	// Async save
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), consts.DurationAuditSave*time.Second)
		defer cancel()
		_ = u.repo.Postgres.Audit.Log.Create(bgCtx, al)
	}()
}

func evaluateConditions(conditions map[string]any, env map[string]any) bool {
	if len(conditions) == 0 {
		return true
	}

	// 1. Handle Logical Operators at top level
	if andBlock, ok := conditions[consts.OpAnd].([]any); ok {
		for _, cond := range andBlock {
			if condMap, ok := cond.(map[string]any); ok {
				if !evaluateConditions(condMap, env) {
					return false
				}
			}
		}
		// All AND conditions met
		if len(conditions) == 1 {
			return true
		}
	}

	if orBlock, ok := conditions[consts.OpOr].([]any); ok {
		match := false
		for _, cond := range orBlock {
			if condMap, ok := cond.(map[string]any); ok {
				if evaluateConditions(condMap, env) {
					match = true
					break
				}
			}
		}
		if !match {
			return false
		}
		if len(conditions) == 1 {
			return true
		}
	}

	if notBlock, ok := conditions[consts.OpNot].(map[string]any); ok {
		if evaluateConditions(notBlock, env) {
			return false
		}
		// If NOT satisfied (meaning inner condition failed), we continue
		if len(conditions) == 1 {
			return true
		}
	}

	// 2. Iterate over standard keys
	for key, expectedVal := range conditions {
		// Skip logical keys processed above
		if key == consts.OpAnd || key == consts.OpOr || key == consts.OpNot {
			continue
		}

		// Handle legacy nested objects (e.g. "target": {...})
		// Kept for backward compat with previous step
		if key == consts.KeyTarget {
			targetEnv, ok := env[consts.KeyTarget].(map[string]any)
			if !ok {
				return false
			}
			targetConditions, ok := expectedVal.(map[string]any)
			if !ok {
				return false
			}
			if !evaluateConditions(targetConditions, targetEnv) {
				return false
			}
			continue
		}

		// Parse "user.role_in", "env.ip_in", etc.
		namespace, attr, op := parseKey(key)

		// Resolve actual value from Environment
		actualVal := resolveEnvValue(env, namespace, attr)

		// Resolve expected value (handle dynamic references like "$user.id")
		resolvedExpected := resolveExpectedValue(expectedVal, env)

		if !checkCondition(actualVal, op, resolvedExpected) {
			return false
		}
	}

	return true
}

// resolveEnvValue fetches value from env like env["user"]["role"]
func resolveEnvValue(env map[string]any, namespace, attr string) any {
	// 1. Try direct lookup (legacy plain keys like "role_id")
	if namespace == "" {
		if val, ok := env[attr]; ok {
			return val
		}
		return nil
	}

	// 2. Try namespaced lookup (e.g. env["user"]["role"])
	if section, ok := env[namespace].(map[string]any); ok {
		if val, ok := section[attr]; ok {
			return val
		}
	} else {
		// Fallback: maybe flattened like "user_id" in root?
		// for "user.id" try "user_id"
		flatKey := namespace + "_" + attr
		if val, ok := env[flatKey]; ok {
			return val
		}
	}

	return nil
}

// resolveExpectedValue handles "$user.region" references
func resolveExpectedValue(val any, env map[string]any) any {
	sVal, ok := val.(string)
	if !ok {
		return val
	}

	if strings.HasPrefix(sVal, "$") {
		refKey := strings.TrimPrefix(sVal, "$") // "user.region"
		ns, attr, _ := parseKey(refKey)         // op is ignored here
		return resolveEnvValue(env, ns, attr)
	}
	return val
}

// parseKey splits "user.age_gte" -> namespace="user", attr="age", op="gte"
// "role_in" -> ns="", attr="role", op="in"
func parseKey(key string) (string, string, string) {
	var namespace, rest string

	// Split namespace (dot)
	parts := strings.SplitN(key, ".", 2)
	if len(parts) == 2 {
		namespace = parts[0]
		rest = parts[1]
	} else {
		namespace = ""
		rest = key
	}

	// Split operator (underscore from end)
	// Heuristic: Check for known operators? Or just last underscore?
	// The prompt implies `age_gte`.
	// Let's protect against "user_id" being split into "user" and "id" as op if "id" was an op.
	// But "id" is not an op. "in", "gte", "equals" are ops.

	i := strings.LastIndex(rest, "_")
	if i == -1 {
		return namespace, rest, consts.OpEquals
	}

	possibleOp := rest[i+1:]
	// basic validation to avoid splitting "user_id" -> attr="user", op="id"
	// List of likely ops from our functionality
	switch possibleOp {
	case "eq", consts.OpEquals, "neq", consts.OpNotEquals, consts.OpIn, consts.OpNotIn, consts.OpContains, consts.OpAny, consts.OpAll, consts.OpGreaterThan, consts.OpLessThan, consts.OpGreaterOrEquals, consts.OpLessOrEquals, consts.OpBetween:
		return namespace, rest[:i], possibleOp
	}

	// No valid operator suffix found, assume whole thing is attribute
	return namespace, rest, consts.OpEquals
}

func checkCondition(actual any, op string, expected any) bool {
	sActual := fmt.Sprintf("%v", actual)
	sExpected := fmt.Sprintf("%v", expected)

	switch op {
	case consts.OpEquals, "eq":
		return sActual == sExpected
	case consts.OpNotEquals, "neq":
		return sActual != sExpected
	case consts.OpIn:
		return isInList(sActual, expected)
	case consts.OpNotIn:
		return !isInList(sActual, expected)
	case consts.OpContains:
		// Check if actual string contains expected substring
		return strings.Contains(sActual, sExpected)
	case consts.OpAny:
		// Check if actual is a list and contains ANY of expected items
		// or actual is list and shares intersection with expected list
		return listIntersects(actual, expected)
	case consts.OpAll:
		// Check if actual is a list and contains ALL of expected items
		return listContainsAll(actual, expected)
	case consts.OpBetween:
		// Handle time or number range
		// expected should be map with "from", "to"
		return checkBetween(actual, expected)
	default:
		// Unsupported operator, fail safe
		return false
	}
}

func isInList(actual string, list any) bool {
	// Retrieve list elements using reflection or type switch if needed,
	// but for "any", JSON unmarshal gives []interface{}
	items, ok := list.([]any)
	if !ok {
		// Try string slice
		if strs, ok := list.([]string); ok {
			for _, s := range strs {
				if s == actual {
					return true
				}
			}
		}
		return false
	}
	for _, item := range items {
		if fmt.Sprintf("%v", item) == actual {
			return true
		}
	}
	return false
}

func listIntersects(actualList any, expectedList any) bool {
	// Convert both to []string for comparison
	actuals := toStringSlice(actualList)
	expecteds := toStringSlice(expectedList)

	for _, a := range actuals {
		for _, e := range expecteds {
			if a == e {
				return true
			}
		}
	}
	return false
}

func listContainsAll(actualList any, expectedList any) bool {
	actuals := toStringSlice(actualList)
	expecteds := toStringSlice(expectedList)

	for _, e := range expecteds {
		found := slices.Contains(actuals, e)
		if !found {
			return false
		}
	}
	return true
}

func toStringSlice(val any) []string {
	var res []string
	if list, ok := val.([]any); ok {
		for _, v := range list {
			res = append(res, fmt.Sprintf("%v", v))
		}
	} else if list, ok := val.([]string); ok {
		res = list
	} else {
		// Single value treated as list of one
		res = append(res, fmt.Sprintf("%v", val))
	}
	return res
}

func checkBetween(actual any, rangeVal any) bool {
	// Simplified for Time (HH:MM) or basic numbers
	// rangeVal should be map[string]any { from, to, timezone? }
	rMap, ok := rangeVal.(map[string]any)
	if !ok {
		return false
	}

	from, _ := rMap["from"].(string)
	to, _ := rMap["to"].(string)

	// Assume time-based comparison for now if parsing succeeds
	// TODO: Add robust date/number handling
	sActual := fmt.Sprintf("%v", actual)

	// String comparison works for ISO timestamps and fixed format "09:00" vs "18:00"
	return sActual >= from && sActual <= to
}
