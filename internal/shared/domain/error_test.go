package domain_test

import (
	"errors"
	"testing"

	"gct/internal/shared/domain"
)

func TestDomainError_Error(t *testing.T) {
	err := domain.NewDomainError("NOT_FOUND", "user not found")

	expected := "NOT_FOUND: user not found"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestDomainError_Code(t *testing.T) {
	err := domain.NewDomainError("VALIDATION", "invalid input")

	if err.Code() != "VALIDATION" {
		t.Errorf("expected code 'VALIDATION', got %q", err.Code())
	}
}

func TestDomainError_Is_MatchByCode(t *testing.T) {
	err1 := domain.NewDomainError("NOT_FOUND", "user not found")
	err2 := domain.NewDomainError("NOT_FOUND", "order not found")

	if !errors.Is(err1, err2) {
		t.Error("expected errors.Is to match by code")
	}
}

func TestDomainError_Is_DifferentCode(t *testing.T) {
	err1 := domain.NewDomainError("NOT_FOUND", "user not found")
	err2 := domain.NewDomainError("VALIDATION", "invalid input")

	if errors.Is(err1, err2) {
		t.Error("expected errors.Is to not match different codes")
	}
}

func TestDomainError_Is_NonDomainError(t *testing.T) {
	err1 := domain.NewDomainError("NOT_FOUND", "user not found")
	err2 := errors.New("some other error")

	if errors.Is(err1, err2) {
		t.Error("expected errors.Is to not match non-DomainError")
	}
}
