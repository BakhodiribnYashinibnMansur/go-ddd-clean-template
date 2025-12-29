package mock

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"

	"gct/internal/domain"
)

// Pagination generates a fake domain.Pagination
func Pagination() *domain.Pagination {
	return &domain.Pagination{
		Limit:  int64(gofakeit.IntRange(1, 100)),
		Offset: int64(gofakeit.IntRange(0, 1000)),
		Total:  int64(gofakeit.IntRange(1, 10000)),
	}
}

// PaginationWithValues generates a domain.Pagination with specific values
func PaginationWithValues(limit, offset, total int64) *domain.Pagination {
	return &domain.Pagination{
		Limit:  limit,
		Offset: offset,
		Total:  total,
	}
}

// Lang generates a fake domain.Lang
func Lang() *domain.Lang {
	return &domain.Lang{
		Uz: gofakeit.Sentence(3),
		Ru: gofakeit.Sentence(3),
		En: gofakeit.Sentence(3),
	}
}

// File generates a fake domain.File
func File() *domain.File {
	return &domain.File{
		Name: gofakeit.FirstName() + gofakeit.LastName() + ".jpg",
		Link: "https://example.com/files/" + uuid.New().String() + ".jpg",
	}
}

// Files generates multiple fake domain.File
func Files(count int) []*domain.File {
	files := make([]*domain.File, count)
	for i := range count {
		files[i] = File()
	}
	return files
}

// Time generates a fake time
func Time() time.Time {
	return gofakeit.Date()
}

// TimeRange generates a fake time within a range
func TimeRange(start, end time.Time) time.Time {
	delta := end.Sub(start)
	randomDelta := time.Duration(gofakeit.IntRange(0, int(delta.Seconds()))) * time.Second
	return start.Add(randomDelta)
}

// FutureTime generates a fake future time
func FutureTime() time.Time {
	return time.Now().Add(time.Duration(gofakeit.IntRange(1, 86400)) * time.Second) // 1 second to 24 hours
}

// PastTime generates a fake past time
func PastTime() time.Time {
	return time.Now().Add(-time.Duration(gofakeit.IntRange(1, 86400)) * time.Second) // 1 second to 24 hours ago
}

// String generates a fake string
func String() string {
	return gofakeit.LetterN(uint(gofakeit.IntRange(5, 20)))
}

// Email generates a fake email
func Email() string {
	return gofakeit.Email()
}

// Phone generates a fake phone number
func Phone() string {
	return gofakeit.Phone()
}

// UUID generates a fake UUID
func UUID() uuid.UUID {
	return uuid.New()
}

// Int generates a fake integer
func Int() int {
	return gofakeit.Int()
}

// Int64 generates a fake int64
func Int64() int64 {
	return int64(gofakeit.Int())
}

// Bool generates a fake boolean
func Bool() bool {
	return gofakeit.Bool()
}
