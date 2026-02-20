package auditlog

import (
	"context"

	"gct/internal/domain"
)

func (uc *UseCase) GetActions(ctx context.Context, in *domain.AuditLogsFilter) ([]domain.ActionEntry, int, error) {
	logs, total, err := uc.Gets(ctx, in)
	if err != nil {
		return nil, 0, err
	}

	data := make([]domain.ActionEntry, len(logs))
	for i, l := range logs {
		actorID := ""
		if l.UserID != nil {
			actorID = l.UserID.String()
		}

		targetID := ""
		if l.ResourceID != nil {
			targetID = l.ResourceID.String()
		}

		ip := ""
		if l.IPAddress != nil {
			ip = *l.IPAddress
		}

		data[i] = domain.ActionEntry{
			ActorID:   actorID,
			Action:    string(l.Action),
			TargetID:  targetID,
			IP:        ip,
			CreatedAt: l.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return data, total, nil
}
