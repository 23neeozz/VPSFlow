package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/libs/go/security/jwt"
	"github.com/vpsflow/vpsflow/services/auth/internal/domain"
	"github.com/vpsflow/vpsflow/services/auth/internal/usecase"
)

type memoryUsers struct {
	users map[string]domain.User
	hash  map[string]string
}

func (m *memoryUsers) CreateUser(ctx context.Context, user domain.User, passwordHash string) error {
	if _, ok := m.users[user.Email]; ok {
		return errors.New(errors.CodeConflict, "email already registered")
	}
	m.users[user.Email] = user
	m.hash[user.Email] = passwordHash
	return nil
}

func (m *memoryUsers) GetUserByEmail(ctx context.Context, email string) (*domain.User, string, error) {
	user, ok := m.users[email]
	if !ok {
		return nil, "", errors.New(errors.CodeNotFound, "user not found")
	}
	return &user, m.hash[email], nil
}

func (m *memoryUsers) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	for _, user := range m.users {
		if user.ID == userID {
			copy := user
			return &copy, nil
		}
	}
	return nil, errors.New(errors.CodeNotFound, "user not found")
}

func (m *memoryUsers) SetMFAEnabled(ctx context.Context, userID string, enabled bool) error {
	return nil
}

type memorySessions struct {
	items map[string]domain.Session
}

func (m *memorySessions) CreateSession(ctx context.Context, session domain.Session) error {
	m.items[session.ID] = session
	return nil
}

func (m *memorySessions) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	session, ok := m.items[sessionID]
	if !ok {
		return nil, errors.New(errors.CodeNotFound, "session not found")
	}
	return &session, nil
}

func (m *memorySessions) RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time) error {
	session, ok := m.items[sessionID]
	if !ok {
		return errors.New(errors.CodeNotFound, "session not found")
	}
	session.RevokedAt = &revokedAt
	m.items[sessionID] = session
	return nil
}

type memoryRefresh struct {
	items map[string]domain.RefreshToken
}

func (m *memoryRefresh) CreateRefreshToken(ctx context.Context, token domain.RefreshToken) error {
	m.items[token.TokenHash] = token
	return nil
}

func (m *memoryRefresh) GetByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	token, ok := m.items[tokenHash]
	if !ok {
		return nil, errors.New(errors.CodeUnauthorized, "refresh token is invalid or expired")
	}
	return &token, nil
}

func (m *memoryRefresh) RevokeToken(ctx context.Context, tokenID string, revokedAt time.Time, replacedBy *string) error {
	for key, token := range m.items {
		if token.ID == tokenID {
			token.RevokedAt = &revokedAt
			token.ReplacedBy = replacedBy
			m.items[key] = token
			return nil
		}
	}
	return nil
}

func (m *memoryRefresh) RevokeFamily(ctx context.Context, familyID string, revokedAt time.Time) error {
	return nil
}

func (m *memoryRefresh) RevokeSessionTokens(ctx context.Context, sessionID string, revokedAt time.Time) error {
	return nil
}

type memoryMFA struct{}

func (m *memoryMFA) GetSecret(ctx context.Context, userID string) (string, bool, error) {
	return "", false, errors.New(errors.CodeNotFound, "mfa secret not found")
}

func (m *memoryMFA) CreateChallenge(ctx context.Context, challenge domain.MFAChallenge) error {
	return nil
}

func (m *memoryMFA) GetChallenge(ctx context.Context, challengeID string) (*domain.MFAChallenge, error) {
	return nil, errors.New(errors.CodeNotFound, "mfa challenge not found")
}

func (m *memoryMFA) MarkChallengeVerified(ctx context.Context, challengeID string, verifiedAt time.Time) error {
	return nil
}

type memoryAttempts struct{}

func (m *memoryAttempts) RecordAttempt(ctx context.Context, email, ipAddress string, success bool, at time.Time) error {
	return nil
}

func (m *memoryAttempts) CountRecentFailures(ctx context.Context, email string, since time.Time) (int, error) {
	return 0, nil
}

func newTestService(t *testing.T) *usecase.Service {
	t.Helper()
	jwtService, err := jwt.NewService(jwt.Config{
		SigningKey: []byte("dev-only-change-in-production-min-32-chars"),
		AccessTTL:  time.Minute,
	})
	if err != nil {
		t.Fatalf("jwt service: %v", err)
	}

	return usecase.NewService(
		&memoryUsers{users: map[string]domain.User{}, hash: map[string]string{}},
		&memorySessions{items: map[string]domain.Session{}},
		&memoryRefresh{items: map[string]domain.RefreshToken{}},
		&memoryMFA{},
		&memoryAttempts{},
		jwtService,
		time.Minute,
		24*time.Hour,
		24*time.Hour,
		5*time.Minute,
		5,
		15*time.Minute,
	)
}

func TestRegisterAndLogin(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	tokens, err := svc.Register(ctx, "user@example.com", "User", "SecurePass123!", usecase.ClientMetadata{})
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("expected tokens from register")
	}

	loginTokens, err := svc.Login(ctx, "user@example.com", "SecurePass123!", usecase.ClientMetadata{})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if loginTokens.AccessToken == "" {
		t.Fatal("expected access token from login")
	}
}

func TestRegisterRejectsWeakPassword(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.Register(context.Background(), "user@example.com", "User", "weak", usecase.ClientMetadata{})
	if err == nil || !errors.IsCode(err, errors.CodeValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}
