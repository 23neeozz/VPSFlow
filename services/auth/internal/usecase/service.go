package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/libs/go/ids"
	"github.com/vpsflow/vpsflow/libs/go/security/jwt"
	"github.com/vpsflow/vpsflow/libs/go/security/password"
	"github.com/vpsflow/vpsflow/libs/go/security/token"
	"github.com/vpsflow/vpsflow/services/auth/internal/domain"
	"github.com/vpsflow/vpsflow/services/auth/internal/port"
	"github.com/pquerna/otp/totp"
)

// ClientMetadata captures request metadata for sessions.
type ClientMetadata struct {
	IPAddress         string
	UserAgent         string
	DeviceFingerprint string
}

// Service orchestrates authentication workflows.
type Service struct {
	users           port.UserRepository
	sessions        port.SessionRepository
	refreshTokens   port.RefreshTokenRepository
	mfa             port.MFARepository
	loginAttempts   port.LoginAttemptRepository
	jwt             *jwt.Service
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	sessionTTL      time.Duration
	mfaChallengeTTL time.Duration
	maxLoginAttempts int
	loginLockout     time.Duration
	now              func() time.Time
}

// NewService creates an authentication use case service.
func NewService(
	users port.UserRepository,
	sessions port.SessionRepository,
	refreshTokens port.RefreshTokenRepository,
	mfa port.MFARepository,
	loginAttempts port.LoginAttemptRepository,
	jwtService *jwt.Service,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration,
	sessionTTL time.Duration,
	mfaChallengeTTL time.Duration,
	maxLoginAttempts int,
	loginLockout time.Duration,
) *Service {
	return &Service{
		users:            users,
		sessions:         sessions,
		refreshTokens:    refreshTokens,
		mfa:              mfa,
		loginAttempts:    loginAttempts,
		jwt:              jwtService,
		accessTokenTTL:   accessTokenTTL,
		refreshTokenTTL:  refreshTokenTTL,
		sessionTTL:       sessionTTL,
		mfaChallengeTTL:  mfaChallengeTTL,
		maxLoginAttempts: maxLoginAttempts,
		loginLockout:     loginLockout,
		now:              time.Now,
	}
}

// Register creates a new user and returns auth tokens.
func (s *Service) Register(ctx context.Context, email, name, plaintextPassword string, meta ClientMetadata) (*domain.AuthTokens, error) {
	email = normalizeEmail(email)
	if email == "" || strings.TrimSpace(name) == "" {
		return nil, errors.New(errors.CodeValidation, "email and name are required")
	}
	if err := password.Validate(plaintextPassword); err != nil {
		return nil, err
	}

	if _, _, err := s.users.GetUserByEmail(ctx, email); err == nil {
		return nil, errors.New(errors.CodeConflict, "email already registered")
	} else if !errors.IsCode(err, errors.CodeNotFound) {
		return nil, err
	}

	hash, err := password.Hash(plaintextPassword)
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to hash password", err)
	}

	userID, err := ids.New("usr")
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to generate user id", err)
	}

	now := s.now().UTC()
	user := domain.User{
		ID:        userID,
		Email:     email,
		Name:      strings.TrimSpace(name),
		Status:    domain.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.users.CreateUser(ctx, user, hash); err != nil {
		return nil, err
	}

	return s.issueSessionTokens(ctx, user, meta)
}

// Login authenticates a user with email and password.
func (s *Service) Login(ctx context.Context, email, plaintextPassword string, meta ClientMetadata) (*domain.AuthTokens, error) {
	email = normalizeEmail(email)
	if email == "" || plaintextPassword == "" {
		return nil, errors.New(errors.CodeValidation, "email and password are required")
	}

	if err := s.ensureLoginAllowed(ctx, email, meta.IPAddress); err != nil {
		return nil, err
	}

	user, passwordHash, err := s.users.GetUserByEmail(ctx, email)
	if err != nil {
		_ = s.loginAttempts.RecordAttempt(ctx, email, meta.IPAddress, false, s.now().UTC())
		if errors.IsCode(err, errors.CodeNotFound) {
			return nil, errors.New(errors.CodeUnauthorized, "invalid email or password")
		}
		return nil, err
	}

	if user.Status != domain.UserStatusActive {
		return nil, errors.New(errors.CodeForbidden, "account is disabled")
	}

	ok, err := password.Verify(passwordHash, plaintextPassword)
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to verify password", err)
	}
	if !ok {
		_ = s.loginAttempts.RecordAttempt(ctx, email, meta.IPAddress, false, s.now().UTC())
		return nil, errors.New(errors.CodeUnauthorized, "invalid email or password")
	}

	_ = s.loginAttempts.RecordAttempt(ctx, email, meta.IPAddress, true, s.now().UTC())

	if user.MFAEnabled {
		return s.createMFAChallenge(ctx, user, meta)
	}

	return s.issueSessionTokens(ctx, *user, meta)
}

// Refresh rotates refresh tokens and issues a new access token.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*domain.AuthTokens, error) {
	if refreshToken == "" {
		return nil, errors.New(errors.CodeValidation, "refresh_token is required")
	}

	now := s.now().UTC()
	tokenHash := token.Hash(refreshToken)
	stored, err := s.refreshTokens.GetByHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if !stored.IsActive(now) {
		_ = s.refreshTokens.RevokeFamily(ctx, stored.FamilyID, now)
		return nil, errors.New(errors.CodeUnauthorized, "refresh token is invalid or expired")
	}

	session, err := s.sessions.GetSession(ctx, stored.SessionID)
	if err != nil {
		return nil, err
	}
	if !session.IsActive(now) {
		return nil, errors.New(errors.CodeUnauthorized, "session is no longer active")
	}

	user, err := s.users.GetUserByID(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}
	if user.Status != domain.UserStatusActive {
		return nil, errors.New(errors.CodeForbidden, "account is disabled")
	}

	newRefreshValue, err := token.GenerateOpaque(32)
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to generate refresh token", err)
	}

	newTokenID, err := ids.New("rtk")
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to generate refresh token id", err)
	}

	newToken := domain.RefreshToken{
		ID:        newTokenID,
		SessionID: stored.SessionID,
		UserID:    stored.UserID,
		TokenHash: token.Hash(newRefreshValue),
		FamilyID:  stored.FamilyID,
		ExpiresAt: now.Add(s.refreshTokenTTL),
		CreatedAt: now,
	}

	if err := s.refreshTokens.CreateRefreshToken(ctx, newToken); err != nil {
		return nil, err
	}

	replacedBy := newTokenID
	if err := s.refreshTokens.RevokeToken(ctx, stored.ID, now, &replacedBy); err != nil {
		return nil, err
	}

	accessToken, expiresAt, err := s.jwt.IssueAccessToken(user.ID, session.ID, user.Email)
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to issue access token", err)
	}

	return &domain.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: newRefreshValue,
		ExpiresIn:    int(expiresAt.Sub(now).Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// Logout revokes the session associated with an access token.
func (s *Service) Logout(ctx context.Context, accessToken string) error {
	claims, err := s.jwt.ParseAccessToken(accessToken)
	if err != nil {
		return err
	}

	now := s.now().UTC()
	if err := s.refreshTokens.RevokeSessionTokens(ctx, claims.SessionID, now); err != nil {
		return err
	}
	return s.sessions.RevokeSession(ctx, claims.SessionID, now)
}

// VerifyMFA completes MFA challenge and issues tokens.
func (s *Service) VerifyMFA(ctx context.Context, challengeID, code string) (*domain.AuthTokens, error) {
	if challengeID == "" || code == "" {
		return nil, errors.New(errors.CodeValidation, "challenge_id and code are required")
	}

	now := s.now().UTC()
	challenge, err := s.mfa.GetChallenge(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	if !challenge.IsPending(now) {
		return nil, errors.New(errors.CodeUnauthorized, "mfa challenge is invalid or expired")
	}

	secret, enabled, err := s.mfa.GetSecret(ctx, challenge.UserID)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, errors.New(errors.CodeUnauthorized, "mfa is not enabled for this account")
	}

	if !totp.Validate(code, secret) {
		return nil, errors.New(errors.CodeUnauthorized, "invalid mfa code")
	}

	if err := s.mfa.MarkChallengeVerified(ctx, challengeID, now); err != nil {
		return nil, err
	}

	user, err := s.users.GetUserByID(ctx, challenge.UserID)
	if err != nil {
		return nil, err
	}

	session, err := s.sessions.GetSession(ctx, challenge.SessionID)
	if err != nil {
		return nil, err
	}
	if !session.IsActive(now) {
		return nil, errors.New(errors.CodeUnauthorized, "session is no longer active")
	}

	accessToken, expiresAt, err := s.jwt.IssueAccessToken(user.ID, session.ID, user.Email)
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to issue access token", err)
	}

	refreshValue, refreshToken, err := s.createRefreshToken(ctx, session.ID, user.ID, now)
	if err != nil {
		return nil, err
	}
	_ = refreshToken

	return &domain.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshValue,
		ExpiresIn:    int(expiresAt.Sub(now).Seconds()),
		TokenType:    "Bearer",
	}, nil
}

// ValidateAccessToken parses and validates an access token.
func (s *Service) ValidateAccessToken(tokenString string) (*jwt.Claims, error) {
	return s.jwt.ParseAccessToken(tokenString)
}

func (s *Service) issueSessionTokens(ctx context.Context, user domain.User, meta ClientMetadata) (*domain.AuthTokens, error) {
	now := s.now().UTC()

	sessionID, err := ids.New("ses")
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to generate session id", err)
	}

	session := domain.Session{
		ID:                sessionID,
		UserID:            user.ID,
		DeviceFingerprint: meta.DeviceFingerprint,
		IPAddress:         meta.IPAddress,
		UserAgent:         meta.UserAgent,
		ExpiresAt:         now.Add(s.sessionTTL),
		CreatedAt:         now,
	}
	if err := s.sessions.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	accessToken, expiresAt, err := s.jwt.IssueAccessToken(user.ID, session.ID, user.Email)
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to issue access token", err)
	}

	refreshValue, _, err := s.createRefreshToken(ctx, session.ID, user.ID, now)
	if err != nil {
		return nil, err
	}

	return &domain.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshValue,
		ExpiresIn:    int(expiresAt.Sub(now).Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func (s *Service) createRefreshToken(ctx context.Context, sessionID, userID string, now time.Time) (string, domain.RefreshToken, error) {
	refreshValue, err := token.GenerateOpaque(32)
	if err != nil {
		return "", domain.RefreshToken{}, errors.Wrap(errors.CodeInternal, "failed to generate refresh token", err)
	}

	tokenID, err := ids.New("rtk")
	if err != nil {
		return "", domain.RefreshToken{}, errors.Wrap(errors.CodeInternal, "failed to generate refresh token id", err)
	}

	familyID, err := ids.New("fam")
	if err != nil {
		return "", domain.RefreshToken{}, errors.Wrap(errors.CodeInternal, "failed to generate token family id", err)
	}

	refreshToken := domain.RefreshToken{
		ID:        tokenID,
		SessionID: sessionID,
		UserID:    userID,
		TokenHash: token.Hash(refreshValue),
		FamilyID:  familyID,
		ExpiresAt: now.Add(s.refreshTokenTTL),
		CreatedAt: now,
	}

	if err := s.refreshTokens.CreateRefreshToken(ctx, refreshToken); err != nil {
		return "", domain.RefreshToken{}, err
	}

	return refreshValue, refreshToken, nil
}

func (s *Service) createMFAChallenge(ctx context.Context, user *domain.User, meta ClientMetadata) (*domain.AuthTokens, error) {
	now := s.now().UTC()

	sessionID, err := ids.New("ses")
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to generate session id", err)
	}

	session := domain.Session{
		ID:                sessionID,
		UserID:            user.ID,
		DeviceFingerprint: meta.DeviceFingerprint,
		IPAddress:         meta.IPAddress,
		UserAgent:         meta.UserAgent,
		ExpiresAt:         now.Add(s.mfaChallengeTTL),
		CreatedAt:         now,
	}
	if err := s.sessions.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	challengeID, err := ids.New("mfa")
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to generate mfa challenge id", err)
	}

	challenge := domain.MFAChallenge{
		ID:        challengeID,
		UserID:    user.ID,
		SessionID: session.ID,
		ExpiresAt: now.Add(s.mfaChallengeTTL),
		CreatedAt: now,
	}
	if err := s.mfa.CreateChallenge(ctx, challenge); err != nil {
		return nil, err
	}

	return &domain.AuthTokens{
		TokenType:   "Bearer",
		MFARequired: true,
		ChallengeID: challengeID,
	}, nil
}

func (s *Service) ensureLoginAllowed(ctx context.Context, email, ipAddress string) error {
	since := s.now().UTC().Add(-s.loginLockout)
	failures, err := s.loginAttempts.CountRecentFailures(ctx, email, since)
	if err != nil {
		return err
	}
	if failures >= s.maxLoginAttempts {
		return errors.New(errors.CodeRateLimited, "too many failed login attempts, try again later")
	}
	_ = ipAddress
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
