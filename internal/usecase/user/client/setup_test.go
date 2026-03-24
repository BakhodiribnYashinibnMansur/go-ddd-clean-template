package client_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"gct/config"
	"gct/internal/domain"
	"gct/internal/repo/persistent"
	"gct/internal/repo/persistent/postgres"
	"gct/internal/repo/persistent/postgres/user"
	clientuc "gct/internal/usecase/user/client"
	"gct/internal/shared/infrastructure/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockClientRepo implements client.RepoI
type MockClientRepo struct {
	mock.Mock
}

func (m *MockClientRepo) Create(ctx context.Context, u *domain.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockClientRepo) Get(ctx context.Context, f *domain.UserFilter) (*domain.User, error) {
	args := m.Called(ctx, f)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockClientRepo) Gets(ctx context.Context, f *domain.UsersFilter) ([]*domain.User, int, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]*domain.User), args.Int(1), args.Error(2)
}

func (m *MockClientRepo) Update(ctx context.Context, u *domain.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockClientRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockClientRepo) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockClientRepo) BulkDeactivate(ctx context.Context, ids []string) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *MockClientRepo) BulkDelete(ctx context.Context, ids []string) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *MockClientRepo) Approve(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockClientRepo) ChangeRole(ctx context.Context, id, role string) error {
	args := m.Called(ctx, id, role)
	return args.Error(0)
}

// MockSessionRepo implements session.RepoI
type MockSessionRepo struct {
	mock.Mock
}

func (m *MockSessionRepo) Create(ctx context.Context, s *domain.Session) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockSessionRepo) Delete(ctx context.Context, f *domain.SessionFilter) error {
	args := m.Called(ctx, f)
	return args.Error(0)
}

func (m *MockSessionRepo) Revoke(ctx context.Context, f *domain.SessionFilter) error {
	args := m.Called(ctx, f)
	return args.Error(0)
}

func (m *MockSessionRepo) Update(ctx context.Context, s *domain.Session) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

func (m *MockSessionRepo) Get(ctx context.Context, f *domain.SessionFilter) (*domain.Session, error) {
	args := m.Called(ctx, f)
	return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *MockSessionRepo) Gets(ctx context.Context, f *domain.SessionsFilter) ([]*domain.Session, int, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]*domain.Session), args.Int(1), args.Error(2)
}

func (m *MockSessionRepo) GetByUser(ctx context.Context, userID int64) ([]*domain.Session, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*domain.Session), args.Error(1)
}

// Setup
func setup(t *testing.T) (clientuc.UseCaseI, *MockClientRepo, *MockSessionRepo) {
	clientRepo := new(MockClientRepo)
	sessionRepo := new(MockSessionRepo)

	// Construct persistent.Repo manually with mocks
	r := &persistent.Repo{
		Postgres: &postgres.Repo{
			User: &user.User{
				Client:      clientRepo,
				SessionRepo: sessionRepo,
			},
		},
	}

	log := logger.New("debug")
	cfg := &config.Config{}
	cfg.JWT.AccessTTL = time.Hour
	cfg.JWT.RefreshTTL = 24 * time.Hour
	cfg.JWT.Issuer = "gct"

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	// Convert key to PEM
	privBytes := x509.MarshalPKCS1PrivateKey(key)
	privPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privBytes,
		},
	)
	cfg.JWT.PrivateKey = string(privPem)

	// Re-create UC so it parses the key
	uc := clientuc.New(r, log, cfg)

	return uc, clientRepo, sessionRepo
}
