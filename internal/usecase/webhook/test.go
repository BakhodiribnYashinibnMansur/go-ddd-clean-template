package webhook

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (uc *UseCase) Test(ctx context.Context, id uuid.UUID) error {
	w, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	payload := `{"event":"test","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, w.URL, strings.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if w.Secret != "" {
		httpReq.Header.Set("X-Webhook-Secret", w.Secret)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("webhook test failed: %w", err)
	}
	defer resp.Body.Close()
	return nil
}
