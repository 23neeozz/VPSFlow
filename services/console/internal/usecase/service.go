package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/vpsflow/vpsflow/libs/go/errors"
	"github.com/vpsflow/vpsflow/libs/go/ids"
	"github.com/vpsflow/vpsflow/services/console/internal/domain"
	"github.com/vpsflow/vpsflow/services/console/internal/port"
)

type Service struct {
	repo       port.Repository
	vmClient   port.VMClient
	tokenTTL   time.Duration
	consoleURL string
}

func NewService(repo port.Repository, vmClient port.VMClient, tokenTTL time.Duration, consoleURL string) *Service {
	if tokenTTL <= 0 {
		tokenTTL = 5 * time.Minute
	}
	return &Service{repo: repo, vmClient: vmClient, tokenTTL: tokenTTL, consoleURL: consoleURL}
}

func (s *Service) CreateSession(ctx context.Context, tenantID, vmID, userID, authorization string) (*domain.SessionToken, error) {
	if tenantID == "" || vmID == "" || userID == "" {
		return nil, errors.New(errors.CodeValidation, "tenant_id, vm_id and user_id are required")
	}

	status, _, _, err := s.vmClient.GetVM(ctx, tenantID, vmID, authorization)
	if err != nil {
		return nil, err
	}
	if status != "running" {
		return nil, errors.New(errors.CodeConflict, "vm must be running to open console")
	}

	sessionID, err := ids.New("cns")
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to generate session id", err)
	}

	rawToken, tokenHash, err := generateToken()
	if err != nil {
		return nil, errors.Wrap(errors.CodeInternal, "failed to generate console token", err)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(s.tokenTTL)
	session := domain.Session{
		ID: sessionID, TenantID: tenantID, VMID: vmID, UserID: userID,
		Status: domain.SessionActive, ExpiresAt: expiresAt, CreatedAt: now,
	}
	if err := s.repo.CreateSession(ctx, session, tokenHash); err != nil {
		return nil, err
	}

	proxyURL := fmt.Sprintf("%s/console/ws?token=%s", s.consoleURL, rawToken)
	return &domain.SessionToken{SessionID: sessionID, Token: rawToken, ExpiresAt: expiresAt, ProxyURL: proxyURL}, nil
}

func (s *Service) ValidateToken(ctx context.Context, rawToken string) (*domain.Session, error) {
	if rawToken == "" {
		return nil, errors.New(errors.CodeValidation, "token is required")
	}
	return s.repo.GetSessionByTokenHash(ctx, hashToken(rawToken))
}

func (s *Service) RevokeSession(ctx context.Context, sessionID string) error {
	return s.repo.RevokeSession(ctx, sessionID)
}

func generateToken() (string, string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	raw := base64.RawURLEncoding.EncodeToString(buf)
	return raw, hashToken(raw), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
