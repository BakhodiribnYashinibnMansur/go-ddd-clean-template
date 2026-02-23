package translation

import (
	"context"
	"encoding/json"

	apperrors "gct/pkg/errors"

	"github.com/google/uuid"
)

func (r *Repo) Upsert(ctx context.Context, entityType string, entityID uuid.UUID, langCode string, data map[string]string) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to marshal translation data")
	}

	sql := `INSERT INTO ` + tableName + ` (entity_type, entity_id, lang_code, data)
VALUES ($1, $2, $3, $4)
ON CONFLICT (entity_type, entity_id, lang_code)
DO UPDATE SET data = translations.data || EXCLUDED.data, updated_at = NOW()`

	_, err = r.pool.Exec(ctx, sql, entityType, entityID, langCode, raw)
	if err != nil {
		return apperrors.HandlePgError(err, tableName, nil)
	}

	return nil
}

