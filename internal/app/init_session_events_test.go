package app

import (
	"context"
	"testing"

	sessiondomain "gct/internal/context/iam/generic/session/domain"
	"gct/internal/kernel/application"
	shareddomain "gct/internal/kernel/domain"
	"gct/internal/context/iam/generic/user"
	usercommand "gct/internal/context/iam/generic/user/application/command"
	userdomain "gct/internal/context/iam/generic/user/domain"

	"github.com/google/uuid"
)

// ---------------------------------------------------------------------------
// Mocks for session event tests
// ---------------------------------------------------------------------------

type sessionTestLogger struct{}

func (l *sessionTestLogger) Debug(args ...any)                                    {}
func (l *sessionTestLogger) Debugf(template string, args ...any)                  {}
func (l *sessionTestLogger) Debugw(msg string, keysAndValues ...any)              {}
func (l *sessionTestLogger) Info(args ...any)                                     {}
func (l *sessionTestLogger) Infof(template string, args ...any)                   {}
func (l *sessionTestLogger) Infow(msg string, keysAndValues ...any)               {}
func (l *sessionTestLogger) Warn(args ...any)                                     {}
func (l *sessionTestLogger) Warnf(template string, args ...any)                   {}
func (l *sessionTestLogger) Warnw(msg string, keysAndValues ...any)               {}
func (l *sessionTestLogger) Error(args ...any)                                    {}
func (l *sessionTestLogger) Errorf(template string, args ...any)                  {}
func (l *sessionTestLogger) Errorw(msg string, keysAndValues ...any)              {}
func (l *sessionTestLogger) Fatal(args ...any)                                    {}
func (l *sessionTestLogger) Fatalf(template string, args ...any)                  {}
func (l *sessionTestLogger) Fatalw(msg string, keysAndValues ...any)              {}
func (l *sessionTestLogger) Debugc(_ context.Context, _ string, _ ...any)         {}
func (l *sessionTestLogger) Infoc(_ context.Context, _ string, _ ...any)          {}
func (l *sessionTestLogger) Warnc(_ context.Context, _ string, _ ...any)          {}
func (l *sessionTestLogger) Errorc(_ context.Context, _ string, _ ...any)         {}
func (l *sessionTestLogger) Fatalc(_ context.Context, _ string, _ ...any)         {}

type sessionTestEventBus struct {
	handlers map[string][]application.EventHandler
}

func newSessionTestEventBus() *sessionTestEventBus {
	return &sessionTestEventBus{
		handlers: make(map[string][]application.EventHandler),
	}
}

func (e *sessionTestEventBus) Publish(ctx context.Context, events ...shareddomain.DomainEvent) error {
	for _, event := range events {
		if handlers, ok := e.handlers[event.EventName()]; ok {
			for _, h := range handlers {
				if err := h(ctx, event); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (e *sessionTestEventBus) Subscribe(eventName string, handler application.EventHandler) error {
	e.handlers[eventName] = append(e.handlers[eventName], handler)
	return nil
}

type sessionTestUserRepo struct {
	signOutCalled     bool
	signOutUserID     uuid.UUID
	signOutSessionID  uuid.UUID
	revokeAllCalled   bool
	revokeAllUserID   uuid.UUID
}

func (m *sessionTestUserRepo) Save(_ context.Context, _ *userdomain.User) error   { return nil }
func (m *sessionTestUserRepo) Update(_ context.Context, _ *userdomain.User) error { return nil }
func (m *sessionTestUserRepo) Delete(_ context.Context, _ uuid.UUID) error        { return nil }
func (m *sessionTestUserRepo) FindByID(_ context.Context, id uuid.UUID) (*userdomain.User, error) {
	return nil, userdomain.ErrUserNotFound
}
func (m *sessionTestUserRepo) List(_ context.Context, _ shareddomain.Pagination) ([]*userdomain.User, int64, error) {
	return nil, 0, nil
}
func (m *sessionTestUserRepo) FindByPhone(_ context.Context, _ userdomain.Phone) (*userdomain.User, error) {
	return nil, userdomain.ErrUserNotFound
}
func (m *sessionTestUserRepo) FindByEmail(_ context.Context, _ userdomain.Email) (*userdomain.User, error) {
	return nil, userdomain.ErrUserNotFound
}
func (m *sessionTestUserRepo) FindDefaultRoleID(_ context.Context) (uuid.UUID, error) {
	return uuid.New(), nil
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestSubscribeSessionEvents_RevokeRequested(t *testing.T) {
	eventBus := newSessionTestEventBus()
	log := &sessionTestLogger{}
	repo := &sessionTestUserRepo{}
	eb := newSessionTestEventBus()

	userBC := &user.BoundedContext{
		SignOut:    usercommand.NewSignOutHandler(repo, eb, log),
		RevokeAll: usercommand.NewRevokeAllSessionsHandler(repo, eb, log),
	}

	subscribeSessionEvents(eventBus, userBC, log)

	// Verify handlers were subscribed
	if len(eventBus.handlers["session.revoke_requested"]) == 0 {
		t.Fatal("expected handler for session.revoke_requested")
	}
	if len(eventBus.handlers["session.revoke_all_requested"]) == 0 {
		t.Fatal("expected handler for session.revoke_all_requested")
	}
}

func TestSubscribeSessionEvents_RevokeEvent_TypeMismatch(t *testing.T) {
	eventBus := newSessionTestEventBus()
	log := &sessionTestLogger{}
	repo := &sessionTestUserRepo{}
	eb := newSessionTestEventBus()

	userBC := &user.BoundedContext{
		SignOut:    usercommand.NewSignOutHandler(repo, eb, log),
		RevokeAll: usercommand.NewRevokeAllSessionsHandler(repo, eb, log),
	}

	subscribeSessionEvents(eventBus, userBC, log)

	// Publish a wrong event type — should not panic, handler returns nil
	wrongEvent := sessiondomain.NewSessionRevokeAllRequested(uuid.New())
	handler := eventBus.handlers["session.revoke_requested"][0]
	err := handler(context.Background(), wrongEvent)
	if err != nil {
		t.Fatalf("expected nil error for type mismatch, got %v", err)
	}
}

func TestSubscribeSessionEvents_RevokeAllEvent_TypeMismatch(t *testing.T) {
	eventBus := newSessionTestEventBus()
	log := &sessionTestLogger{}
	repo := &sessionTestUserRepo{}
	eb := newSessionTestEventBus()

	userBC := &user.BoundedContext{
		SignOut:    usercommand.NewSignOutHandler(repo, eb, log),
		RevokeAll: usercommand.NewRevokeAllSessionsHandler(repo, eb, log),
	}

	subscribeSessionEvents(eventBus, userBC, log)

	// Publish a wrong event type — should not panic
	wrongEvent := sessiondomain.NewSessionRevokeRequested(uuid.New(), uuid.New())
	handler := eventBus.handlers["session.revoke_all_requested"][0]
	err := handler(context.Background(), wrongEvent)
	if err != nil {
		t.Fatalf("expected nil error for type mismatch, got %v", err)
	}
}
