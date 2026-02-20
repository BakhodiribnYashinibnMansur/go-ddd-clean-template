package auditlog

import (
	"context"

	"gct/internal/domain"
)

func (uc *UseCase) GetSessions(ctx context.Context, in *domain.AuditLogsFilter) ([]domain.SessionEntry, int, error) {
	logs, total, err := uc.Gets(ctx, in)
	if err != nil {
		return nil, 0, err
	}

	data := make([]domain.SessionEntry, len(logs))
	for i, l := range logs {
		uid := ""
		if l.UserID != nil {
			uid = l.UserID.String()
		}

		sid := ""
		if l.SessionID != nil {
			sid = l.SessionID.String()
		}

		ip := ""
		if l.IPAddress != nil {
			ip = *l.IPAddress
		}

		data[i] = domain.SessionEntry{
			SessionID: sid,
			UserID:    uid,
			Event:     string(l.Action),
			Source:    "user",
			IP:        ip,
			CreatedAt: l.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return data, total, nil
}
