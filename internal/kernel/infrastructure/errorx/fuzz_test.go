package errorx

import "testing"

func FuzzGetSeverity(f *testing.F) {
	f.Add("BAD_REQUEST")
	f.Add("UNAUTHORIZED")
	f.Add("INTERNAL_ERROR")
	f.Add("TIMEOUT")
	f.Add("")
	f.Add("UNKNOWN_CODE_XYZ")

	f.Fuzz(func(t *testing.T, code string) {
		_ = GetSeverity(code)
	})
}

func FuzzGetCategory(f *testing.F) {
	f.Add("BAD_REQUEST")
	f.Add("UNAUTHORIZED")
	f.Add("NOT_FOUND")
	f.Add("INTERNAL_ERROR")
	f.Add("")
	f.Add("UNKNOWN_CODE_XYZ")

	f.Fuzz(func(t *testing.T, code string) {
		_ = GetCategory(code)
	})
}

func FuzzGetLayer(f *testing.F) {
	f.Add("HANDLER_ERROR")
	f.Add("SERVICE_ERROR")
	f.Add("REPO_NOT_FOUND")
	f.Add("EXT_TIMEOUT")
	f.Add("")
	f.Add("UNKNOWN")

	f.Fuzz(func(t *testing.T, code string) {
		_ = GetLayer(code)
	})
}

func FuzzIsRetryable(f *testing.F) {
	f.Add("TIMEOUT")
	f.Add("BAD_REQUEST")
	f.Add("REPO_TIMEOUT")
	f.Add("")
	f.Add("UNKNOWN_CODE_XYZ")

	f.Fuzz(func(t *testing.T, code string) {
		_ = IsRetryable(code)
	})
}

func FuzzGetUserMessage(f *testing.F) {
	f.Add("BAD_REQUEST", "en")
	f.Add("UNAUTHORIZED", "uz")
	f.Add("INTERNAL_ERROR", "ru")
	f.Add("", "en")
	f.Add("UNKNOWN_CODE", "")
	f.Add("NOT_FOUND", "fr")

	f.Fuzz(func(t *testing.T, code, lang string) {
		_ = GetUserMessage(code, lang)
	})
}
