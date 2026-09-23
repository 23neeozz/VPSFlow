package port

import (
	"context"
	"time"

	"github.com/bosscloud/bosscloud/services/auth/internal/domain"
)

// UserRepository persists user identities.
type UserRepository interface {
	CreateUser(ctx context.Context, user domain.User, passwordHash string) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, string, error)
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)
	SetMFAEnabled(ctx context.Context, userID string, enabled bool) error
}

// SessionRepository manages sessions.
type SessionRepository interface {
	CreateSession(ctx context.Context, session domain.Session) error
	GetSession(ctx context.Context, sessionID string) (*domain.Session, error)
	RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time) error
}

// RefreshTokenRepository manages refresh token rotation.
type RefreshTokenRepository interface {
	CreateRefreshToken(ctx context.Context, token domain.RefreshToken) error
	GetByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	RevokeToken(ctx context.Context, tokenID string, revokedAt time.Time, replacedBy *string) error
	RevokeFamily(ctx context.Context, familyID string, revokedAt time.Time) error
	RevokeSessionTokens(ctx context.Context, sessionID string, revokedAt time.Time) error
}

// MFARepository stores MFA secrets and challenges.
type MFARepository interface {
	GetSecret(ctx context.Context, userID string) (secret string, enabled bool, err error)
	CreateChallenge(ctx context.Context, challenge domain.MFAChallenge) error
	GetChallenge(ctx context.Context, challengeID string) (*domain.MFAChallenge, error)
	MarkChallengeVerified(ctx context.Context, challengeID string, verifiedAt time.Time) error
}

// LoginAttemptRepository tracks authentication attempts.
type LoginAttemptRepository interface {
	RecordAttempt(ctx context.Context, email, ipAddress string, success bool, at time.Time) error
	CountRecentFailures(ctx context.Context, email string, since time.Time) (int, error)
}
