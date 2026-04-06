package entity_test

import (
	"testing"

	"gct/internal/context/iam/generic/user/domain/entity"
	"gct/internal/context/iam/generic/user/domain/service"
)

// ---------------------------------------------------------------------------
// Benchmark: Password Hashing (bcrypt is intentionally slow)
// ---------------------------------------------------------------------------

func BenchmarkNewPasswordFromRaw(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = entity.NewPasswordFromRaw("BenchmarkP@ss1")
	}
}

func BenchmarkPasswordCompare_Success(b *testing.B) {
	pw, _ := entity.NewPasswordFromRaw("BenchmarkP@ss1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pw.Compare("BenchmarkP@ss1")
	}
}

func BenchmarkPasswordCompare_Failure(b *testing.B) {
	pw, _ := entity.NewPasswordFromRaw("BenchmarkP@ss1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pw.Compare("WrongPassword1")
	}
}

// ---------------------------------------------------------------------------
// Benchmark: Value Object creation
// ---------------------------------------------------------------------------

func BenchmarkNewPhone(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = entity.NewPhone("+998901234567")
	}
}

func BenchmarkNewEmail(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = entity.NewEmail("bench@example.com")
	}
}

// ---------------------------------------------------------------------------
// Benchmark: User Aggregate creation
// ---------------------------------------------------------------------------

func BenchmarkNewUser(b *testing.B) {
	phone, _ := entity.NewPhone("+998901234567")
	pw, _ := entity.NewPasswordFromRaw("BenchmarkP@ss1")
	email, _ := entity.NewEmail("bench@example.com")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = entity.NewUser(phone, pw, entity.WithEmail(email), entity.WithUsername("bench"))
	}
}

// ---------------------------------------------------------------------------
// Benchmark: Session operations
// ---------------------------------------------------------------------------

func BenchmarkAddSession(b *testing.B) {
	phone, _ := entity.NewPhone("+998901234567")
	pw, _ := entity.NewPasswordFromRaw("BenchmarkP@ss1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		u, _ := entity.NewUser(phone, pw)
		_, _ = u.AddSession(entity.DeviceDesktop, "1.2.3.4", "BenchAgent", "gct-client")
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	phone, _ := entity.NewPhone("+998901234567")
	pw, _ := entity.NewPasswordFromRaw("BenchmarkP@ss1")
	u, _ := entity.NewUser(phone, pw)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = u.VerifyPassword("BenchmarkP@ss1")
	}
}

// ---------------------------------------------------------------------------
// Benchmark: SignInService
// ---------------------------------------------------------------------------

func BenchmarkSignInService(b *testing.B) {
	phone, _ := entity.NewPhone("+998901234567")
	pw, _ := entity.NewPasswordFromRaw("BenchmarkP@ss1")
	svc := &service.SignInService{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		u, _ := entity.NewUser(phone, pw)
		u.Approve()
		_, _ = svc.SignIn(u, "BenchmarkP@ss1", entity.DeviceDesktop, "10.0.0.1", "BenchAgent", "gct-client")
	}
}
