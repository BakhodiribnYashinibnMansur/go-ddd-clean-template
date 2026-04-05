package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedAuditLogs(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding audit logs...", zap.Int("count", count))

	// Get user IDs
	userRows, err := s.pool.Query(ctx, "SELECT id FROM users LIMIT 50")
	if err != nil {
		return fmt.Errorf("failed to get users for audit logs: %w", err)
	}
	defer userRows.Close()

	var userIDs []uuid.UUID
	for userRows.Next() {
		var id uuid.UUID
		if err := userRows.Scan(&id); err != nil {
			return fmt.Errorf("failed to scan user id: %w", err)
		}
		userIDs = append(userIDs, id)
	}

	// Get session IDs
	sessionRows, err := s.pool.Query(ctx, "SELECT id FROM session LIMIT 50")
	if err != nil {
		s.logger.Warnc(ctx, "No sessions found, audit logs will have nil session_id", zap.Error(err))
	}

	var sessionIDs []uuid.UUID
	if sessionRows != nil {
		for sessionRows.Next() {
			var id uuid.UUID
			if err := sessionRows.Scan(&id); err != nil {
				break
			}
			sessionIDs = append(sessionIDs, id)
		}
		sessionRows.Close()
	}

	// Get policy IDs
	policyRows, err := s.pool.Query(ctx, "SELECT id FROM policy LIMIT 50")
	if err != nil {
		s.logger.Warnc(ctx, "No policies found, audit logs will have nil policy_id", zap.Error(err))
	}

	var policyIDs []uuid.UUID
	if policyRows != nil {
		for policyRows.Next() {
			var id uuid.UUID
			if err := policyRows.Scan(&id); err != nil {
				break
			}
			policyIDs = append(policyIDs, id)
		}
		policyRows.Close()
	}

	actions := []string{
		"LOGIN", "LOGOUT", "SESSION_REVOKE",
		"PASSWORD_CHANGE", "ACCESS_GRANTED", "ACCESS_DENIED",
		"POLICY_MATCHED", "POLICY_DENIED",
		"USER_CREATE", "USER_UPDATE", "USER_DELETE",
		"ROLE_ASSIGN", "ROLE_REMOVE", "POLICY_EVALUATED", "ADMIN_CHANGE",
	}
	platforms := []string{"admin", "web", "mobile", "api"}
	resourceTypes := []string{"user", "role", "permission", "policy", "session", "file", "integration"}
	decisions := []string{"ALLOW", "DENY"}
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15",
		"Mozilla/5.0 (Linux; Android 13) AppleWebKit/537.36",
		"curl/8.1.2",
		"PostmanRuntime/7.32.3",
	}

	for i := 0; i < count; i++ {
		var userID *uuid.UUID
		if len(userIDs) > 0 {
			u := userIDs[gofakeit.Number(0, len(userIDs)-1)]
			userID = &u
		}

		var sessionID *uuid.UUID
		if len(sessionIDs) > 0 {
			sid := sessionIDs[gofakeit.Number(0, len(sessionIDs)-1)]
			sessionID = &sid
		}

		var policyID *uuid.UUID
		if len(policyIDs) > 0 && gofakeit.Bool() {
			pid := policyIDs[gofakeit.Number(0, len(policyIDs)-1)]
			policyID = &pid
		}

		action := actions[gofakeit.Number(0, len(actions)-1)]
		resourceType := resourceTypes[gofakeit.Number(0, len(resourceTypes)-1)]
		resourceID := uuid.New()
		platform := platforms[gofakeit.Number(0, len(platforms)-1)]
		ip := fmt.Sprintf("%d.%d.%d.%d", gofakeit.Number(1, 223), gofakeit.Number(0, 255), gofakeit.Number(0, 255), gofakeit.Number(1, 254))
		ua := userAgents[gofakeit.Number(0, len(userAgents)-1)]
		decision := decisions[gofakeit.Number(0, 1)]
		success := decision == "ALLOW"
		var errorMsg *string
		if !success {
			msg := "Access denied: insufficient permissions"
			errorMsg = &msg
		}
		createdAt := gofakeit.DateRange(time.Now().AddDate(0, -1, 0), time.Now())

		_, err := s.pool.Exec(ctx,
			`INSERT INTO audit_log (id, user_id, session_id, action, resource_type, resource_id, platform, ip_address, user_agent, policy_id, decision, success, error_message, created_at)
			 VALUES ($1, $2, $3, $4::audit_action_type, $5, $6, $7, $8::inet, $9, $10, $11, $12, $13, $14)`,
			uuid.New(), userID, sessionID, action, resourceType, resourceID, platform, ip, ua, policyID, decision, success, errorMsg, createdAt,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create audit log", zap.Error(err), zap.String("action", action))
		}
	}

	return nil
}
