package domain_test

import (
	"testing"
	"time"

	domain "gct/internal/context/iam/generic/user/domain"

	"github.com/google/uuid"
)

func TestDomainEvents_TableDriven(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	oldRoleID := uuid.New()
	newRoleID := uuid.New()

	tests := []struct {
		name          string
		event         interface{ EventName() string }
		expectedName  string
		checkTime     bool
		checkAggrID   bool
		aggregateID   uuid.UUID
	}{
		{
			name:         "UserCreated",
			event:        domain.NewUserCreated(userID, "+998901234567"),
			expectedName: "user.created",
			checkTime:    true,
			checkAggrID:  true,
			aggregateID:  userID,
		},
		{
			name:         "UserSignedIn",
			event:        domain.NewUserSignedIn(userID, sessionID, "10.0.0.1"),
			expectedName: "user.signed_in",
			checkTime:    true,
			checkAggrID:  true,
			aggregateID:  userID,
		},
		{
			name:         "UserDeactivated",
			event:        domain.NewUserDeactivated(userID),
			expectedName: "user.deactivated",
			checkTime:    true,
			checkAggrID:  true,
			aggregateID:  userID,
		},
		{
			name:         "PasswordChanged",
			event:        domain.NewPasswordChanged(userID),
			expectedName: "user.password_changed",
			checkTime:    true,
			checkAggrID:  true,
			aggregateID:  userID,
		},
		{
			name:         "UserApproved",
			event:        domain.NewUserApproved(userID),
			expectedName: "user.approved",
			checkTime:    true,
			checkAggrID:  true,
			aggregateID:  userID,
		},
		{
			name:         "RoleChanged",
			event:        domain.NewRoleChanged(userID, &oldRoleID, newRoleID),
			expectedName: "user.role_changed",
			checkTime:    true,
			checkAggrID:  true,
			aggregateID:  userID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.event.EventName() != tt.expectedName {
				t.Errorf("expected event name %q, got %q", tt.expectedName, tt.event.EventName())
			}

			// All our events implement DomainEvent with OccurredAt and AggregateID
			type domainEvent interface {
				EventName() string
				OccurredAt() time.Time
				AggregateID() uuid.UUID
			}

			de, ok := tt.event.(domainEvent)
			if !ok {
				t.Fatal("event does not implement DomainEvent")
			}

			if tt.checkTime {
				if de.OccurredAt().IsZero() {
					t.Error("occurredAt should not be zero")
				}
				if time.Since(de.OccurredAt()) > time.Second {
					t.Error("occurredAt should be recent")
				}
			}

			if tt.checkAggrID {
				if de.AggregateID() != tt.aggregateID {
					t.Errorf("expected aggregateID %s, got %s", tt.aggregateID, de.AggregateID())
				}
			}
		})
	}
}

func TestRoleChanged_CarriesOldAndNewRoleIDs(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	oldRole := uuid.New()
	newRole := uuid.New()

	event := domain.NewRoleChanged(userID, &oldRole, newRole)

	if event.OldRoleID == nil || *event.OldRoleID != oldRole {
		t.Error("expected old role ID to be set")
	}
	if event.NewRoleID != newRole {
		t.Error("expected new role ID to be set")
	}
}

func TestRoleChanged_NilOldRole(t *testing.T) {
	t.Parallel()

	event := domain.NewRoleChanged(uuid.New(), nil, uuid.New())

	if event.OldRoleID != nil {
		t.Error("expected old role ID to be nil")
	}
}

func TestUserSignedIn_CarriesSessionAndIP(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	sessionID := uuid.New()

	event := domain.NewUserSignedIn(userID, sessionID, "192.168.1.1")

	if event.SessionID != sessionID {
		t.Errorf("expected sessionID %s, got %s", sessionID, event.SessionID)
	}
	if event.IPAddress != "192.168.1.1" {
		t.Errorf("expected IP 192.168.1.1, got %s", event.IPAddress)
	}
}

func TestUserCreated_CarriesPhone(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	event := domain.NewUserCreated(userID, "+998901234567")

	if event.Phone != "+998901234567" {
		t.Errorf("expected phone +998901234567, got %s", event.Phone)
	}
}
