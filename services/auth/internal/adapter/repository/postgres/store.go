package postgres

import (
	"context"
	"errors"
	"time"

	apperrors "github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/libs/go/ids"
	"github.com/vpsflow/vpsflow/services/auth/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store provides PostgreSQL-backed auth persistence.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore creates a PostgreSQL auth store.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) CreateUser(ctx context.Context, user domain.User, passwordHash string) error {
	credentialID, err := ids.New("cred")
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to generate credential id", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to begin transaction", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO users (user_id, email, name, email_verified, mfa_enabled, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, user.ID, user.Email, user.Name, user.EmailVerified, user.MFAEnabled, user.Status, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return apperrors.New(apperrors.CodeConflict, "email already registered")
		}
		return apperrors.Wrap(apperrors.CodeInternal, "failed to create user", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO credentials (credential_id, user_id, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, credentialID, user.ID, passwordHash, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to create credential", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to commit user creation", err)
	}

	return nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*domain.User, string, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT u.user_id, u.email, u.name, u.email_verified, u.mfa_enabled, u.status, u.created_at, u.updated_at, c.password_hash
		FROM users u
		INNER JOIN credentials c ON c.user_id = u.user_id
		WHERE u.email = $1
	`, email)

	user, passwordHash, err := scanUserWithPassword(row)
	if err != nil {
		return nil, "", err
	}
	return user, passwordHash, nil
}

func (s *Store) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT user_id, email, name, email_verified, mfa_enabled, status, created_at, updated_at
		FROM users
		WHERE user_id = $1
	`, userID)

	user, err := scanUser(row)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Store) SetMFAEnabled(ctx context.Context, userID string, enabled bool) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE users SET mfa_enabled = $2, updated_at = NOW() WHERE user_id = $1
	`, userID, enabled)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to update mfa status", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.New(apperrors.CodeNotFound, "user not found")
	}
	return nil
}

func (s *Store) CreateSession(ctx context.Context, session domain.Session) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sessions (session_id, user_id, device_fingerprint, ip_address, user_agent, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, session.ID, session.UserID, session.DeviceFingerprint, session.IPAddress, session.UserAgent, session.ExpiresAt, session.CreatedAt)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to create session", err)
	}
	return nil
}

func (s *Store) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT session_id, user_id, device_fingerprint, ip_address, user_agent, revoked_at, expires_at, created_at
		FROM sessions
		WHERE session_id = $1
	`, sessionID)

	var session domain.Session
	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.DeviceFingerprint,
		&session.IPAddress,
		&session.UserAgent,
		&session.RevokedAt,
		&session.ExpiresAt,
		&session.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(apperrors.CodeNotFound, "session not found")
		}
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to get session", err)
	}
	return &session, nil
}

func (s *Store) RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE sessions SET revoked_at = $2 WHERE session_id = $1 AND revoked_at IS NULL
	`, sessionID, revokedAt)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to revoke session", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.New(apperrors.CodeNotFound, "session not found")
	}
	return nil
}

func (s *Store) CreateRefreshToken(ctx context.Context, token domain.RefreshToken) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (token_id, session_id, user_id, token_hash, family_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, token.ID, token.SessionID, token.UserID, token.TokenHash, token.FamilyID, token.ExpiresAt, token.CreatedAt)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to create refresh token", err)
	}
	return nil
}

func (s *Store) GetByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT token_id, session_id, user_id, token_hash, family_id, replaced_by, revoked_at, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`, tokenHash)

	var refreshToken domain.RefreshToken
	err := row.Scan(
		&refreshToken.ID,
		&refreshToken.SessionID,
		&refreshToken.UserID,
		&refreshToken.TokenHash,
		&refreshToken.FamilyID,
		&refreshToken.ReplacedBy,
		&refreshToken.RevokedAt,
		&refreshToken.ExpiresAt,
		&refreshToken.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(apperrors.CodeUnauthorized, "refresh token is invalid or expired")
		}
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to get refresh token", err)
	}
	return &refreshToken, nil
}

func (s *Store) RevokeToken(ctx context.Context, tokenID string, revokedAt time.Time, replacedBy *string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = $2, replaced_by = $3
		WHERE token_id = $1 AND revoked_at IS NULL
	`, tokenID, revokedAt, replacedBy)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to revoke refresh token", err)
	}
	return nil
}

func (s *Store) RevokeFamily(ctx context.Context, familyID string, revokedAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = $2
		WHERE family_id = $1 AND revoked_at IS NULL
	`, familyID, revokedAt)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to revoke token family", err)
	}
	return nil
}

func (s *Store) RevokeSessionTokens(ctx context.Context, sessionID string, revokedAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = $2
		WHERE session_id = $1 AND revoked_at IS NULL
	`, sessionID, revokedAt)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to revoke session refresh tokens", err)
	}
	return nil
}

func (s *Store) GetSecret(ctx context.Context, userID string) (string, bool, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT secret_encrypted, enabled FROM mfa_secrets WHERE user_id = $1
	`, userID)

	var secret string
	var enabled bool
	err := row.Scan(&secret, &enabled)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, apperrors.New(apperrors.CodeNotFound, "mfa secret not found")
		}
		return "", false, apperrors.Wrap(apperrors.CodeInternal, "failed to get mfa secret", err)
	}
	return secret, enabled, nil
}

func (s *Store) CreateChallenge(ctx context.Context, challenge domain.MFAChallenge) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO mfa_challenges (challenge_id, user_id, session_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, challenge.ID, challenge.UserID, challenge.SessionID, challenge.ExpiresAt, challenge.CreatedAt)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to create mfa challenge", err)
	}
	return nil
}

func (s *Store) GetChallenge(ctx context.Context, challengeID string) (*domain.MFAChallenge, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT challenge_id, user_id, session_id, expires_at, verified_at, created_at
		FROM mfa_challenges
		WHERE challenge_id = $1
	`, challengeID)

	var challenge domain.MFAChallenge
	err := row.Scan(
		&challenge.ID,
		&challenge.UserID,
		&challenge.SessionID,
		&challenge.ExpiresAt,
		&challenge.VerifiedAt,
		&challenge.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(apperrors.CodeNotFound, "mfa challenge not found")
		}
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to get mfa challenge", err)
	}
	return &challenge, nil
}

func (s *Store) MarkChallengeVerified(ctx context.Context, challengeID string, verifiedAt time.Time) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE mfa_challenges SET verified_at = $2 WHERE challenge_id = $1 AND verified_at IS NULL
	`, challengeID, verifiedAt)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to verify mfa challenge", err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.New(apperrors.CodeNotFound, "mfa challenge not found")
	}
	return nil
}

func (s *Store) RecordAttempt(ctx context.Context, email, ipAddress string, success bool, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO login_attempts (email, ip_address, success, created_at)
		VALUES ($1, $2, $3, $4)
	`, email, ipAddress, success, at)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to record login attempt", err)
	}
	return nil
}

func (s *Store) CountRecentFailures(ctx context.Context, email string, since time.Time) (int, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM login_attempts
		WHERE email = $1 AND success = FALSE AND created_at >= $2
	`, email, since)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, apperrors.Wrap(apperrors.CodeInternal, "failed to count login failures", err)
	}
	return count, nil
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.EmailVerified,
		&user.MFAEnabled,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.New(apperrors.CodeNotFound, "user not found")
		}
		return nil, apperrors.Wrap(apperrors.CodeInternal, "failed to scan user", err)
	}
	return &user, nil
}

func scanUserWithPassword(row pgx.Row) (*domain.User, string, error) {
	var user domain.User
	var passwordHash string
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.EmailVerified,
		&user.MFAEnabled,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&passwordHash,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", apperrors.New(apperrors.CodeNotFound, "user not found")
		}
		return nil, "", apperrors.Wrap(apperrors.CodeInternal, "failed to scan user", err)
	}
	return &user, passwordHash, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
