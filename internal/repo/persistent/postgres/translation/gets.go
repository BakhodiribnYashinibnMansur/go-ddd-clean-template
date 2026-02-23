package translation

import (
	"context"
	"encoding/json"
	"fmt"

	"gct/internal/domain"
	apperrors "gct/pkg/errors"

	"github.com/Masterminds/squirrel"
)

func (r *Repo) Gets(ctx context.Context, filter domain.TranslationFilter) ([]*domain.Translation, error) {
	query := r.builder.
		Select("id", "entity_type", "entity_id", "lang_code", "data", "created_at", "updated_at").
		From(tableName).
		Where(squirrel.Eq{
			"entity_type": filter.EntityType,
			"entity_id":   filter.EntityID,
		})

	if filter.LangCode != nil {
		query = query.Where(squirrel.Eq{"lang_code": *filter.LangCode})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to build select query")
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}
	defer rows.Close()

	var result []*domain.Translation
	for rows.Next() {
		var t domain.Translation
		var rawData []byte
		if err := rows.Scan(&t.ID, &t.EntityType, &t.EntityID, &t.LangCode, &rawData, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, fmt.Sprintf("failed to scan row: %v", err))
		}
		if err := json.Unmarshal(rawData, &t.Data); err != nil {
			return nil, apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to unmarshal translation data")
		}
		result = append(result, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.HandlePgError(err, tableName, nil)
	}

	return result, nil
}
