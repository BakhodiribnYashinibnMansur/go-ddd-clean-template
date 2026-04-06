package event_test

import (
	"testing"
	"time"

	userevent "gct/internal/context/iam/generic/user/domain/event"

	"github.com/google/uuid"
)

func TestDomainEvents_TableDriven(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	oldRoleID := uuid.New()
	newRoleID := uuid.New()

	tests := []struct {
		name         string
		event        interface{ EventName() string }
		expectedName string
		checkTime    bool
		checkAggrID  bool
		aggregateID  uuid.UUID
	}{
		{
			name:         "UserCreated",
			event:        userevent.NewUserCreated(userID, "+998901234567"),
			expectedName: "user.created",
			checkTime:    true,
			checkAggrID:  true,
			aggregateID:  userID,
		},
		{
			name:         "UserSignedIn",
			event:        userevent.NewUserSignedIn(userID, sessionID, "10.0.0.1"),
			expectedName: "user.signed_in",
			checkTime:    true,
			checkAggrID:  true,
			aggregateID:  userID,
		},
		{
			name:         "UserDeactivated",
			event:        userevent.NewUserDeactivated(userID),
			expectedName: "user.deactivated",
			checkTime:    true,
			checkAggrID:  true,
			aggregateID:  userID,
		},
		{
			name:         "PasswordChanged",
			event:        userevent.NewPasswordChanged(userID),
			expectedName: "user.password_changed",
			checkTime:    true,
			checkAggrID:  true,
			aggregateID:  userID,
		},
		{
			name:         "UserApproved",
			event:        userevent.NewUserApproved(userID),
			expectedName: "user.approved",
			checkTime:    true,
			checkAggrID:  true,
			aggregateID:  userID,
		},
		{
			name:         "RoleChanged",
			event:        userevent.NewRoleChanged(userID, &oldRoleID, newRoleID),
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

	e := userevent.NewRoleChanged(userID, &oldRole, newRole)

	if e.OldRoleID == nil || *e.OldRoleID != oldRole {
		t.Error("expected old role ID to be set")
	}
	if e.NewRoleID != newRole {
		t.Error("expected new role ID to be set")
	}
}

func TestRoleChanged_NilOldRole(t *testing.T) {
	t.Parallel()

	e := userevent.NewRoleChanged(uuid.New(), nil, uuid.New())

	if e.OldRoleID != nil {
		t.Error("expected old role ID to be nil")
	}
}

func TestUserSignedIn_CarriesSessionAndIP(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	sessionID := uuid.New()

	e := userevent.NewUserSignedIn(userID, sessionID, "192.168.1.1")

	if e.SessionID != sessionID {
		t.Errorf("expected sessionID %s, got %s", sessionID, e.SessionID)
	}
	if e.IPAddress != "192.168.1.1" {
		t.Errorf("expected IP 192.168.1.1, got %s", e.IPAddress)
	}
}

func TestUserCreated_CarriesPhone(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	e := userevent.NewUserCreated(userID, "+998901234567")

	if e.Phone != "+998901234567" {
		t.Errorf("expected phone +998901234567, got %s", e.Phone)
	}
}
