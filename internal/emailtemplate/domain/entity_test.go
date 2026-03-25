package domain_test

import (
	"testing"

	"gct/internal/emailtemplate/domain"
)

func TestNewEmailTemplate(t *testing.T) {
	vars := []string{"name", "email"}
	et := domain.NewEmailTemplate("welcome", "Welcome!", "<h1>Hi</h1>", "Hi", vars)

	if et.Name() != "welcome" {
		t.Fatalf("expected name welcome, got %s", et.Name())
	}
	if et.Subject() != "Welcome!" {
		t.Fatalf("expected subject Welcome!, got %s", et.Subject())
	}
	if et.HTMLBody() != "<h1>Hi</h1>" {
		t.Fatalf("expected htmlBody <h1>Hi</h1>, got %s", et.HTMLBody())
	}
	if et.TextBody() != "Hi" {
		t.Fatalf("expected textBody Hi, got %s", et.TextBody())
	}
	if len(et.Variables()) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(et.Variables()))
	}
	if et.ID().String() == "" {
		t.Fatal("expected non-empty ID")
	}
}

func TestEmailTemplate_UpdateDetails(t *testing.T) {
	et := domain.NewEmailTemplate("old", "Old Subject", "<p>old</p>", "old", nil)

	newName := "new"
	newSubject := "New Subject"
	et.UpdateDetails(&newName, &newSubject, nil, nil, nil)

	if et.Name() != "new" {
		t.Fatalf("expected name new, got %s", et.Name())
	}
	if et.Subject() != "New Subject" {
		t.Fatalf("expected subject New Subject, got %s", et.Subject())
	}
	if len(et.Events()) != 1 {
		t.Fatalf("expected 1 event, got %d", len(et.Events()))
	}
	if et.Events()[0].EventName() != "emailtemplate.updated" {
		t.Fatalf("expected event emailtemplate.updated, got %s", et.Events()[0].EventName())
	}
}
