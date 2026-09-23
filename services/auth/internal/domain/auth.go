package domain

import "time"

const (
	UserStatusActive   = "active"
	UserStatusDisabled = "disabled"
)

// User represents a global VPSFlow identity.
type User struct {
	ID            string
	Email         string
	Name          string
	EmailVerified bool
	MFAEnabled    bool
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Session represents an authenticated user session.
type Session struct {
	ID                string
	UserID            string
	DeviceFingerprint string
	IPAddress         string
	UserAgent         string
	RevokedAt         *time.Time
	ExpiresAt         time.Time
	CreatedAt         time.Time
}

func (s Session) IsActive(now time.Time) bool {
	if s.RevokedAt != nil {
		return false
	}
	return now.Before(s.ExpiresAt)
}

// RefreshToken represents a rotatable refresh token.
type RefreshToken struct {
	ID         string
	SessionID  string
	UserID     string
	TokenHash  string
	FamilyID   string
	ReplacedBy *string
	RevokedAt  *time.Time
	ExpiresAt  time.Time
	CreatedAt  time.Time
}

func (t RefreshToken) IsActive(now time.Time) bool {
	if t.RevokedAt != nil {
		return false
	}
	return now.Before(t.ExpiresAt)
}

// MFAChallenge represents a pending MFA verification.
type MFAChallenge struct {
	ID         string
	UserID     string
	SessionID  string
	ExpiresAt  time.Time
	VerifiedAt *time.Time
	CreatedAt  time.Time
}

func (c MFAChallenge) IsPending(now time.Time) bool {
	if c.VerifiedAt != nil {
		return false
	}
	return now.Before(c.ExpiresAt)
}

// AuthTokens is the response payload for successful authentication.
type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	TokenType    string
	MFARequired  bool
	ChallengeID  string
}
