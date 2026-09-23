package jwt

import (
	"fmt"
	"time"

	"github.com/vpsflow/vpsflow/libs/go/errors"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

// Claims represents VPSFlow JWT access token claims.
type Claims struct {
	UserID    string `json:"sub"`
	SessionID string `json:"sid"`
	Email     string `json:"email"`
	jwtlib.RegisteredClaims
}

// Config holds JWT signing configuration.
type Config struct {
	SigningKey []byte
	Issuer     string
	Audience   string
	AccessTTL  time.Duration
}

// Service issues and validates JWT access tokens.
type Service struct {
	cfg Config
}

// NewService creates a JWT service.
func NewService(cfg Config) (*Service, error) {
	if len(cfg.SigningKey) < 32 {
		return nil, fmt.Errorf("signing key must be at least 32 bytes")
	}
	if cfg.Issuer == "" {
		cfg.Issuer = "vpsflow"
	}
	if cfg.Audience == "" {
		cfg.Audience = "vpsflow-api"
	}
	if cfg.AccessTTL <= 0 {
		cfg.AccessTTL = 15 * time.Minute
	}
	return &Service{cfg: cfg}, nil
}

// IssueAccessToken creates a signed JWT access token.
func (s *Service) IssueAccessToken(userID, sessionID, email string) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.cfg.AccessTTL)

	claims := Claims{
		UserID:    userID,
		SessionID: sessionID,
		Email:     email,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    s.cfg.Issuer,
			Audience:  jwtlib.ClaimStrings{s.cfg.Audience},
			Subject:   userID,
			ExpiresAt: jwtlib.NewNumericDate(expiresAt),
			IssuedAt:  jwtlib.NewNumericDate(now),
			NotBefore: jwtlib.NewNumericDate(now),
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.cfg.SigningKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}

	return signed, expiresAt, nil
}

// ParseAccessToken validates and parses an access token.
func (s *Service) ParseAccessToken(tokenString string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &Claims{}, func(token *jwtlib.Token) (any, error) {
		if token.Method != jwtlib.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.cfg.SigningKey, nil
	})
	if err != nil {
		return nil, errors.Wrap(errors.CodeUnauthorized, "invalid access token", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New(errors.CodeUnauthorized, "invalid access token claims")
	}

	return claims, nil
}
