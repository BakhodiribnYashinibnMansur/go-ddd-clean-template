package errors

import (
	"strings"
	"testing"
)

func TestGetAllErrors(t *testing.T) {
	defs := GetAllErrors()
	if len(defs) == 0 {
		t.Fatal("expected non-empty error definitions")
	}

	// Verify some known errors exist
	foundBadRequest := false
	foundNotFound := false
	foundInternal := false
	for _, d := range defs {
		switch d.Code {
		case ErrBadRequest:
			foundBadRequest = true
			if d.HTTPStatus != 400 {
				t.Errorf("ErrBadRequest: expected HTTP status 400, got %d", d.HTTPStatus)
			}
			if d.Layer != "General" {
				t.Errorf("ErrBadRequest: expected layer 'General', got %q", d.Layer)
			}
		case ErrNotFound:
			foundNotFound = true
			if d.HTTPStatus != 404 {
				t.Errorf("ErrNotFound: expected HTTP status 404, got %d", d.HTTPStatus)
			}
		case ErrInternal:
			foundInternal = true
			if d.HTTPStatus != 500 {
				t.Errorf("ErrInternal: expected HTTP status 500, got %d", d.HTTPStatus)
			}
		}
	}

	if !foundBadRequest {
		t.Error("expected to find ErrBadRequest in definitions")
	}
	if !foundNotFound {
		t.Error("expected to find ErrNotFound in definitions")
	}
	if !foundInternal {
		t.Error("expected to find ErrInternal in definitions")
	}
}

func TestGetAllErrors_ContainsAllLayers(t *testing.T) {
	defs := GetAllErrors()

	layers := map[string]bool{}
	for _, d := range defs {
		layers[d.Layer] = true
	}

	expectedLayers := []string{"General", "Repository", "Service", "Handler"}
	for _, l := range expectedLayers {
		if !layers[l] {
			t.Errorf("expected layer %q to be present", l)
		}
	}
}

func TestGetErrorsByFilter(t *testing.T) {
	tests := []struct {
		name     string
		layer    string
		category string
		code     string
		minCount int
	}{
		{
			name:     "filter by layer Repository",
			layer:    "Repository",
			category: "",
			code:     "",
			minCount: 1,
		},
		{
			name:     "filter by category Security",
			layer:    "",
			category: "Security",
			code:     "",
			minCount: 1,
		},
		{
			name:     "filter by code partial match",
			layer:    "",
			category: "",
			code:     "NOT_FOUND",
			minCount: 1,
		},
		{
			name:     "filter by layer and category",
			layer:    "General",
			category: "Validation",
			code:     "",
			minCount: 1,
		},
		{
			name:     "no match returns empty",
			layer:    "NonExistentLayer",
			category: "",
			code:     "",
			minCount: 0,
		},
		{
			name:     "case insensitive layer",
			layer:    "general",
			category: "",
			code:     "",
			minCount: 1,
		},
		{
			name:     "case insensitive category",
			layer:    "",
			category: "security",
			code:     "",
			minCount: 1,
		},
		{
			name:     "case insensitive code partial match",
			layer:    "",
			category: "",
			code:     "not_found",
			minCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetErrorsByFilter(tt.layer, tt.category, tt.code)

			if len(result) < tt.minCount {
				t.Errorf("expected at least %d results, got %d", tt.minCount, len(result))
			}

			// Verify filter was applied
			for _, d := range result {
				if tt.layer != "" && !strings.EqualFold(d.Layer, tt.layer) {
					t.Errorf("expected layer %q, got %q", tt.layer, d.Layer)
				}
				if tt.category != "" && !strings.EqualFold(d.Category, tt.category) {
					t.Errorf("expected category %q, got %q", tt.category, d.Category)
				}
				if tt.code != "" && !strings.Contains(strings.ToLower(d.Code), strings.ToLower(tt.code)) {
					t.Errorf("expected code to contain %q, got %q", tt.code, d.Code)
				}
			}
		})
	}
}

func TestGetNumericCode_AllLayers(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{"bad request", ErrBadRequest, CodeBadRequest},
		{"not found", ErrNotFound, CodeNotFound},
		{"internal", ErrInternal, CodeInternal},
		{"repo not found", ErrRepoNotFound, CodeRepoNotFound},
		{"service not found", ErrServiceNotFound, CodeServiceNotFound},
		{"handler bad request", ErrHandlerBadRequest, CodeHandlerBadRequest},
		{"unknown code returns 9999", "NONEXISTENT_CODE", "9999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetNumericCode(tt.code)
			if got != tt.want {
				t.Errorf("GetNumericCode(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestGetUserMessage_Languages(t *testing.T) {
	tests := []struct {
		name string
		code string
		lang string
	}{
		{"english bad request", ErrBadRequest, "en"},
		{"uzbek bad request", ErrBadRequest, "uz"},
		{"russian bad request", ErrBadRequest, "ru"},
		{"english not found", ErrNotFound, "en"},
		{"unknown code", "NONEXISTENT", "en"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := GetUserMessage(tt.code, tt.lang)
			if msg == "" {
				t.Error("expected non-empty user message")
			}
		})
	}
}

func TestGetUserMessageWithDetails(t *testing.T) {
	t.Run("with details in english", func(t *testing.T) {
		msg := GetUserMessageWithDetails(ErrBadRequest, "en", "field X is required")
		if !strings.Contains(msg, "Details: field X is required") {
			t.Errorf("expected message to contain details, got %q", msg)
		}
	})

	t.Run("with details in uzbek", func(t *testing.T) {
		msg := GetUserMessageWithDetails(ErrBadRequest, "uz", "field X")
		if !strings.Contains(msg, "Tafsilotlar: field X") {
			t.Errorf("expected uzbek details prefix, got %q", msg)
		}
	})

	t.Run("with details in russian", func(t *testing.T) {
		msg := GetUserMessageWithDetails(ErrBadRequest, "ru", "field X")
		if !strings.Contains(msg, "Детали: field X") {
			t.Errorf("expected russian details prefix, got %q", msg)
		}
	})

	t.Run("without details", func(t *testing.T) {
		msg := GetUserMessageWithDetails(ErrBadRequest, "en", "")
		if strings.Contains(msg, "Details:") {
			t.Errorf("expected no details section, got %q", msg)
		}
	})
}
