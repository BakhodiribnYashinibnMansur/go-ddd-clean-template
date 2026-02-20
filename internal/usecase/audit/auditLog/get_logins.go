package auditlog

import (
	"context"

	"gct/internal/domain"
)

func (uc *UseCase) GetLogins(ctx context.Context, in *domain.AuditLogsFilter) ([]domain.LoginEntry, int, error) {
	logs, total, err := uc.Gets(ctx, in)
	if err != nil {
		return nil, 0, err
	}

	data := make([]domain.LoginEntry, len(logs))
	for i, l := range logs {
		status := "success"
		if !l.Success {
			status = "failed"
		}

		reason := ""
		if l.ErrorMessage != nil {
			reason = *l.ErrorMessage
		}

		uid := ""
		if l.UserID != nil {
			uid = l.UserID.String()
		}

		ip := ""
		if l.IPAddress != nil {
			ip = *l.IPAddress
		}

		data[i] = domain.LoginEntry{
			UserID:    uid,
			Status:    status,
			IP:        ip,
			Reason:    reason,
			CreatedAt: l.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return data, total, nil
}
