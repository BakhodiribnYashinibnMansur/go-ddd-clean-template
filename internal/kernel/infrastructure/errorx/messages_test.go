package errorx

import (
	"sync"
	"testing"
)

func TestGetUserMessage_English(t *testing.T) {
	got := GetUserMessage(ErrBadRequest, "en")
	want := "The request could not be processed due to invalid data."
	if got != want {
		t.Errorf("GetUserMessage(%q, \"en\") = %q, want %q", ErrBadRequest, got, want)
	}
}

func TestGetUserMessage_Uzbek(t *testing.T) {
	got := GetUserMessage(ErrBadRequest, "uz")
	want := "So'rov noto'g'ri ma'lumotlar tufayli qayta ishlanmadi."
	if got != want {
		t.Errorf("GetUserMessage(%q, \"uz\") = %q, want %q", ErrBadRequest, got, want)
	}
}

func TestGetUserMessage_Russian(t *testing.T) {
	got := GetUserMessage(ErrBadRequest, "ru")
	want := "Запрос не может быть обработан из-за неверных данных."
	if got != want {
		t.Errorf("GetUserMessage(%q, \"ru\") = %q, want %q", ErrBadRequest, got, want)
	}
}

func TestGetUserMessage_UnknownCode(t *testing.T) {
	got := GetUserMessage("UNKNOWN_CODE_TEST_12345", "en")
	want := "An error occurred. Please try again."
	if got != want {
		t.Errorf("GetUserMessage(unknown, \"en\") = %q, want %q", got, want)
	}
}

func TestGetUserMessage_DefaultsToEnglish(t *testing.T) {
	got := GetUserMessage(ErrBadRequest, "fr")
	want := "The request could not be processed due to invalid data."
	if got != want {
		t.Errorf("GetUserMessage(%q, \"fr\") = %q, want English fallback %q", ErrBadRequest, got, want)
	}
}

func TestGetUserMessageFallback(t *testing.T) {
	tests := []struct {
		name string
		lang string
		want string
	}{
		{"English fallback", "en", "An error occurred. Please try again."},
		{"Uzbek fallback", "uz", "Xatolik yuz berdi. Iltimos, qayta urinib ko'ring."},
		{"Russian fallback", "ru", "Произошла ошибка. Пожалуйста, попробуйте снова."},
		{"Unknown lang fallback", "de", "An error occurred. Please try again."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getUserMessageFallback(tt.lang)
			if got != tt.want {
				t.Errorf("getUserMessageFallback(%q) = %q, want %q", tt.lang, got, tt.want)
			}
		})
	}
}

func TestSetAndGetHTTPStatus(t *testing.T) {
	code := "TEST_CODE_HTTP_STATUS"
	defer RemoveError(code)

	SetHTTPStatus(code, 418)
	got := GetHTTPStatus(code)
	if got != 418 {
		t.Errorf("GetHTTPStatus(%q) = %d, want 418", code, got)
	}
}

func TestGetHTTPStatus_NotSet(t *testing.T) {
	got := GetHTTPStatus("TEST_CODE_NOT_SET_999")
	if got != 0 {
		t.Errorf("GetHTTPStatus for unset code = %d, want 0", got)
	}
}

func TestConfigureError(t *testing.T) {
	code := "TEST_CODE_CONFIGURE"
	defer RemoveError(code)

	config := ErrorDetailConfig{
		Message: UserMessage{
			En: "Test English",
			Uz: "Test Uzbek",
			Ru: "Test Russian",
		},
		HTTPStatus: 422,
	}
	ConfigureError(code, config)

	if got := GetHTTPStatus(code); got != 422 {
		t.Errorf("HTTP status = %d, want 422", got)
	}
	if got := GetUserMessage(code, "en"); got != "Test English" {
		t.Errorf("English message = %q, want %q", got, "Test English")
	}
	if got := GetUserMessage(code, "uz"); got != "Test Uzbek" {
		t.Errorf("Uzbek message = %q, want %q", got, "Test Uzbek")
	}
	if got := GetUserMessage(code, "ru"); got != "Test Russian" {
		t.Errorf("Russian message = %q, want %q", got, "Test Russian")
	}
}

func TestRemoveError(t *testing.T) {
	code := "TEST_CODE_REMOVE"

	SetHTTPStatus(code, 500)
	UpdateUserMessage(code, UserMessage{En: "To be removed"})

	RemoveError(code)

	if got := GetHTTPStatus(code); got != 0 {
		t.Errorf("HTTP status after remove = %d, want 0", got)
	}
	fallback := getUserMessageFallback("en")
	if got := GetUserMessage(code, "en"); got != fallback {
		t.Errorf("message after remove = %q, want fallback %q", got, fallback)
	}
}

func TestGetUserMessageWithDetails_Languages(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		lang    string
		details string
		wantSfx string
	}{
		{
			name:    "English with details",
			code:    ErrBadRequest,
			lang:    "en",
			details: "field X is missing",
			wantSfx: " Details: field X is missing",
		},
		{
			name:    "Uzbek with details",
			code:    ErrBadRequest,
			lang:    "uz",
			details: "X maydoni yo'q",
			wantSfx: " Tafsilotlar: X maydoni yo'q",
		},
		{
			name:    "Russian with details",
			code:    ErrBadRequest,
			lang:    "ru",
			details: "поле X отсутствует",
			wantSfx: " Детали: поле X отсутствует",
		},
		{
			name:    "empty details returns base message",
			code:    ErrBadRequest,
			lang:    "en",
			details: "",
			wantSfx: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := GetUserMessage(tt.code, tt.lang)
			got := GetUserMessageWithDetails(tt.code, tt.lang, tt.details)
			want := base + tt.wantSfx
			if got != want {
				t.Errorf("GetUserMessageWithDetails() = %q, want %q", got, want)
			}
		})
	}
}

func TestUpdateUserMessage(t *testing.T) {
	code := "TEST_CODE_UPDATE"
	defer RemoveError(code)

	original := UserMessage{En: "Original", Uz: "Asl", Ru: "Оригинал"}
	UpdateUserMessage(code, original)

	if got := GetUserMessage(code, "en"); got != "Original" {
		t.Errorf("before update: got %q, want %q", got, "Original")
	}

	updated := UserMessage{En: "Updated", Uz: "Yangilangan", Ru: "Обновлено"}
	UpdateUserMessage(code, updated)

	if got := GetUserMessage(code, "en"); got != "Updated" {
		t.Errorf("after update EN: got %q, want %q", got, "Updated")
	}
	if got := GetUserMessage(code, "uz"); got != "Yangilangan" {
		t.Errorf("after update UZ: got %q, want %q", got, "Yangilangan")
	}
	if got := GetUserMessage(code, "ru"); got != "Обновлено" {
		t.Errorf("after update RU: got %q, want %q", got, "Обновлено")
	}
}

func TestConcurrentAccess(t *testing.T) {
	const goroutines = 50
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines * 3)

	// Writers: SetHTTPStatus + UpdateUserMessage
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			code := "TEST_CODE_CONCURRENT"
			for j := 0; j < iterations; j++ {
				SetHTTPStatus(code, 400+id)
				UpdateUserMessage(code, UserMessage{
					En: "concurrent test",
					Uz: "bir vaqtda test",
					Ru: "параллельный тест",
				})
			}
		}(i)
	}

	// Readers: GetUserMessage
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			code := "TEST_CODE_CONCURRENT"
			for j := 0; j < iterations; j++ {
				_ = GetUserMessage(code, "en")
				_ = GetHTTPStatus(code)
			}
		}()
	}

	// Mixed: ConfigureError + RemoveError
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			code := "TEST_CODE_CONCURRENT_MIX"
			for j := 0; j < iterations; j++ {
				ConfigureError(code, ErrorDetailConfig{
					Message:    UserMessage{En: "mix"},
					HTTPStatus: 500,
				})
				_ = GetUserMessage(code, "en")
				RemoveError(code)
			}
		}(i)
	}

	wg.Wait()

	// Clean up
	RemoveError("TEST_CODE_CONCURRENT")
	RemoveError("TEST_CODE_CONCURRENT_MIX")
}
