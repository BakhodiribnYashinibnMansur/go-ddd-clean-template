package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (s *Seeder) seedFileMetadata(ctx context.Context, count int) error {
	s.logger.Infoc(ctx, "Seeding file metadata...", zap.Int("count", count))

	// Get user IDs for uploaded_by reference
	rows, err := s.pool.Query(ctx, "SELECT id FROM users LIMIT 50")
	if err != nil {
		return fmt.Errorf("failed to get users for file metadata: %w", err)
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("failed to scan user id: %w", err)
		}
		userIDs = append(userIDs, id)
	}

	now := time.Now()

	fileTypes := []struct {
		ext      string
		mimeType string
		minSize  int64
		maxSize  int64
	}{
		{"pdf", "application/pdf", 50000, 10000000},
		{"png", "image/png", 10000, 5000000},
		{"jpg", "image/jpeg", 20000, 8000000},
		{"xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", 30000, 15000000},
		{"csv", "text/csv", 1000, 50000000},
		{"docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", 20000, 10000000},
		{"json", "application/json", 100, 1000000},
		{"txt", "text/plain", 100, 500000},
	}

	for i := 0; i < count; i++ {
		ft := fileTypes[gofakeit.Number(0, len(fileTypes)-1)]
		originalName := fmt.Sprintf("%s_%s.%s", gofakeit.Word(), gofakeit.Word(), ft.ext)
		storedName := fmt.Sprintf("%s.%s", uuid.New().String(), ft.ext)
		size := int64(gofakeit.Number(int(ft.minSize), int(ft.maxSize)))
		url := fmt.Sprintf("/files/%s/%s", "uploads", storedName)

		var uploadedBy *uuid.UUID
		if len(userIDs) > 0 {
			u := userIDs[gofakeit.Number(0, len(userIDs)-1)]
			uploadedBy = &u
		}

		_, err := s.pool.Exec(ctx,
			`INSERT INTO file_metadata (id, original_name, stored_name, bucket, url, size, mime_type, uploaded_by, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			uuid.New(), originalName, storedName, "uploads", url, size, ft.mimeType, uploadedBy, now, now,
		)
		if err != nil {
			s.logger.Warnc(ctx, "Failed to create file metadata", zap.Error(err), zap.String("name", originalName))
		}
	}

	return nil
}
