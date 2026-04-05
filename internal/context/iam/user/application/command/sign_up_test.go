package command

import (
	"context"
	"testing"
)

func TestSignUpHandler_Handle(t *testing.T) {
	repo := &mockUserRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewSignUpHandler(repo, eventBus, log)

	username := "newuser"
	email := "newuser@example.com"
	cmd := SignUpCommand{
		Phone:    "+998901234567",
		Password: "StrongP@ss123",
		Username: &username,
		Email:    &email,
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo.savedUser == nil {
		t.Fatal("expected user to be saved")
	}

	if repo.savedUser.Phone().Value() != "+998901234567" {
		t.Errorf("expected phone +998901234567, got %s", repo.savedUser.Phone().Value())
	}

	if !repo.savedUser.IsApproved() {
		t.Error("sign-up user should be auto-approved")
	}

	if repo.savedUser.Email() == nil || repo.savedUser.Email().Value() != "newuser@example.com" {
		t.Error("expected email to be set")
	}

	if repo.savedUser.Username() == nil || *repo.savedUser.Username() != "newuser" {
		t.Error("expected username to be set")
	}

	if len(eventBus.publishedEvents) == 0 {
		t.Fatal("expected events to be published")
	}

	if eventBus.publishedEvents[0].EventName() != "user.created" {
		t.Errorf("expected user.created event, got %s", eventBus.publishedEvents[0].EventName())
	}
}

func TestSignUpHandler_MinimalFields(t *testing.T) {
	repo := &mockUserRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewSignUpHandler(repo, eventBus, log)

	cmd := SignUpCommand{
		Phone:    "+998907654321",
		Password: "AnotherP@ss1",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if repo.savedUser == nil {
		t.Fatal("expected user to be saved")
	}

	if repo.savedUser.Email() != nil {
		t.Error("email should be nil when not provided")
	}

	if repo.savedUser.Username() != nil {
		t.Error("username should be nil when not provided")
	}
}

func TestSignUpHandler_InvalidPhone(t *testing.T) {
	repo := &mockUserRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewSignUpHandler(repo, eventBus, log)

	cmd := SignUpCommand{
		Phone:    "bad-phone",
		Password: "StrongP@ss123",
	}

	err := handler.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error for invalid phone")
	}

	if repo.savedUser != nil {
		t.Error("no user should be saved for invalid phone")
	}
}

func TestSignUpHandler_WeakPassword(t *testing.T) {
	repo := &mockUserRepository{}
	eventBus := &mockEventBus{}
	log := &mockLogger{}

	handler := NewSignUpHandler(repo, eventBus, log)

	cmd := SignUpCommand{
		Phone:    "+998901234567",
		Password: "short",
	}

	err := handler.Handle(context.Background(), cmd)
	if err == nil {
		t.Fatal("expected error for weak password")
	}
}
